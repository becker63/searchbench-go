package eino

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

type proposalPayload struct {
	ArtifactID   string   `json:"artifact_id"`
	ArtifactName string   `json:"artifact_name"`
	InterfaceID  string   `json:"interface_id"`
	Code         string   `json:"code"`
	Summary      string   `json:"summary"`
	RiskNotes    []string `json:"risk_notes"`
}

func finalizeProposal(raw string, target pureoptimizer.Target, attemptNumber int) (*pureoptimizer.Proposal, *pureoptimizer.Failure) {
	var payload proposalPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return nil, &pureoptimizer.Failure{
			Phase:     pureoptimizer.PhaseFinalizeProposal,
			Kind:      pureoptimizer.FailureKindPolicyProposalFailed,
			Message:   "parse optimizer proposal JSON",
			Cause:     err,
			Attempt:   attemptNumber,
			Retryable: true,
		}
	}

	if strings.TrimSpace(payload.ArtifactID) == "" {
		return nil, invalidProposalFailure("proposal artifact_id is required", attemptNumber)
	}
	if payload.ArtifactID != target.OutputArtifactID.String() {
		return nil, invalidProposalFailure(
			fmt.Sprintf("proposal artifact_id %q does not match target %q", payload.ArtifactID, target.OutputArtifactID),
			attemptNumber,
		)
	}
	if payload.ArtifactName != target.OutputName {
		return nil, invalidProposalFailure(
			fmt.Sprintf("proposal artifact_name %q does not match target %q", payload.ArtifactName, target.OutputName),
			attemptNumber,
		)
	}
	if payload.InterfaceID != target.InterfaceID {
		return nil, invalidProposalFailure(
			fmt.Sprintf("proposal interface_id %q does not match target %q", payload.InterfaceID, target.InterfaceID),
			attemptNumber,
		)
	}
	if strings.TrimSpace(payload.Code) == "" {
		return nil, invalidProposalFailure("proposal code is required", attemptNumber)
	}
	if strings.Contains(payload.Code, "```") {
		return nil, invalidProposalFailure("proposal code must not contain markdown fences", attemptNumber)
	}
	if filepath.IsAbs(payload.ArtifactName) || strings.Contains(filepath.ToSlash(filepath.Clean(payload.ArtifactName)), "../") {
		return nil, invalidProposalFailure("proposal artifact_name must be relative and must not contain '..'", attemptNumber)
	}
	if !strings.Contains(payload.Code, "def score(") {
		return nil, invalidProposalFailure("proposal code must define def score(...)", attemptNumber)
	}

	proposal := &pureoptimizer.Proposal{
		ArtifactID:   target.OutputArtifactID,
		ArtifactName: payload.ArtifactName,
		InterfaceID:  payload.InterfaceID,
		Code:         payload.Code,
		Summary:      strings.TrimSpace(payload.Summary),
		RiskNotes:    append([]string(nil), payload.RiskNotes...),
	}
	return proposal, nil
}

func invalidProposalFailure(message string, attemptNumber int) *pureoptimizer.Failure {
	return &pureoptimizer.Failure{
		Phase:     pureoptimizer.PhaseFinalizeProposal,
		Kind:      pureoptimizer.FailureKindPolicyProposalFailed,
		Message:   message,
		Cause:     errors.New(message),
		Attempt:   attemptNumber,
		Retryable: true,
	}
}
