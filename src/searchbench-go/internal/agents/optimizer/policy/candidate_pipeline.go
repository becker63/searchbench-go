package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	execpipeline "github.com/becker63/searchbench-go/internal/adapters/pipeline/exec"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/buckdescriptor"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/localpath"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/materialize"
	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// ValidateProposalInWorkspace runs the IC policy pipeline with cwd set to candidate.Root.
// The supplied workspace is used as-is; no hidden rematerialization occurs.
func ValidateProposalInWorkspace(
	ctx context.Context,
	candidate pureoptimizer.ICCandidateWorkspace,
	proposal pureoptimizer.NextChallengerProposal,
) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	if err := candidate.Validate(); err != nil {
		return infraFailure(proposal, fmt.Errorf("candidate workspace: %w", err))
	}
	if strings.TrimSpace(proposal.Code) == "" {
		return pipelineFailureResult(stagePolicyFailure("policy code is empty"))
	}
	if strings.Contains(proposal.Code, "```") {
		return pipelineFailureResult(stagePolicyFailure("policy code contains markdown code fences; emit raw Python only"))
	}
	return validateProposalInWorkspaceRoot(ctx, candidate, proposal)
}

// ValidateProposalWithLocalPathSeed materializes IC from a local path seed, then validates in that workspace.
func ValidateProposalWithLocalPathSeed(
	ctx context.Context,
	proposal pureoptimizer.NextChallengerProposal,
) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	icRoot, err := ResolveIterativeContextRoot()
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("resolve iterative-context root: %w", err))
	}
	seed, err := localpath.Provider{Source: icRoot}.PrepareSeed(ctx)
	if err != nil {
		return infraFailure(proposal, err)
	}
	return validateProposalWithSeed(ctx, seed, proposal)
}

// ValidateProposalWithBuckDescriptor resolves a Buck optimizable backend descriptor seed, then validates.
func ValidateProposalWithBuckDescriptor(
	ctx context.Context,
	proposal pureoptimizer.NextChallengerProposal,
) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	provider := buckdescriptor.Provider{
		DescriptorTarget: "//src/iterative-context:optimizable_backend",
	}
	seed, err := provider.PrepareSeed(ctx)
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("buck descriptor seed: %w", err))
	}
	return validateProposalWithSeed(ctx, seed, proposal)
}

func validateProposalWithSeed(
	ctx context.Context,
	seed pureoptimizer.WorkspaceSeed,
	proposal pureoptimizer.NextChallengerProposal,
) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	mat := materialize.CandidateMaterializer{}
	candidate, cleanup, err := mat.Materialize(seed)
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("materialize candidate workspace: %w", err))
	}
	defer func() { _ = cleanup() }()
	return ValidateProposalInWorkspace(ctx, candidate, proposal)
}

func validateProposalInWorkspaceRoot(
	ctx context.Context,
	candidate pureoptimizer.ICCandidateWorkspace,
	proposal pureoptimizer.NextChallengerProposal,
) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	workspace := filepath.Clean(candidate.Root)
	validationDir := filepath.Join(workspace, "validation")
	if err := os.MkdirAll(validationDir, 0o755); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "create validation directory",
			Cause:   err,
		}
	}

	policyPath := filepath.Join(workspace, proposal.ArtifactName)
	if err := os.WriteFile(policyPath, []byte(proposal.Code), 0o644); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "write staged policy proposal",
			Cause:   err,
		}
	}

	sum := sha256.Sum256([]byte(proposal.Code))
	meta := map[string]any{
		"artifact_id":   proposal.ArtifactID.String(),
		"artifact_name": proposal.ArtifactName,
		"interface_id":  proposal.InterfaceID,
		"symbol":        CanonicalICPolicySymbol,
		"sha256":        hex.EncodeToString(sum[:]),
		"seed_provider": candidate.Seed.Provider,
		"seed_source":   candidate.Seed.Source,
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "encode staged policy metadata",
			Cause:   err,
		}
	}
	if err := os.WriteFile(filepath.Join(workspace, "policy.json"), metaBytes, 0o644); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "write policy.json metadata",
			Cause:   err,
		}
	}

	stageStep := pipeline.StepResult{
		Name:    "stage_policy",
		Command: []string{},
		CWD:     workspace,
		Passed:  true,
	}
	if err := os.WriteFile(filepath.Join(validationDir, "stage-policy.json"), metaBytes, 0o644); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "write stage-policy.json",
			Cause:   err,
		}
	}
	results := []pipeline.StepResult{stageStep}

	uvBin, err := execpipeline.LookupUvBinary()
	if err != nil {
		return infraFailure(proposal, err)
	}

	policyAbs, err := filepath.Abs(policyPath)
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("abs policy path: %w", err))
	}
	wsAbs, err := filepath.Abs(workspace)
	if err != nil {
		return infraFailure(proposal, fmt.Errorf("abs workspace path: %w", err))
	}

	pythonBin, err := resolvePython()
	if err != nil {
		step := pipeline.StepResult{
			Name:                "py_compile",
			Command:             []string{"python3", "-m", "py_compile", proposal.ArtifactName},
			CWD:                 workspace,
			ExitCode:            -1,
			InfrastructureError: err,
		}
		results = append(results, step)
		classification := pipeline.Classify(results)
		return failureFromClassification(results, &classification, proposal)
	}

	preSpec := pipeline.CommandSpec{
		Name:    "py_compile",
		Command: []string{pythonBin, "-m", "py_compile", proposal.ArtifactName},
		CWD:     workspace,
		Timeout: 30 * time.Second,
	}
	preResult := execpipeline.ExecCommandRunner{}.Run(ctx, preSpec)
	preResult.Name = preSpec.Name
	preResult.Command = append([]string(nil), preSpec.Command...)
	preResult.CWD = preSpec.CWD
	if preResult.InfrastructureError == nil && preResult.ExitCode == 0 {
		preResult.Passed = true
	}
	results = append(results, preResult)
	if preResult.Failed() {
		classification := pipeline.Classify(results)
		return failureFromClassification(results, &classification, proposal)
	}

	policyID := proposal.ArtifactID.String()
	icArgv := []string{
		uvBin, "run", "python", "-m", "iterative_context.validate_policy",
		"--policy-path", policyAbs,
		"--policy-id", policyID,
		"--symbol", CanonicalICPolicySymbol,
		"--json",
	}
	allow := execpipeline.ICOptimizerAllowlist{
		UvBinary:         uvBin,
		WorkspaceRootAbs: wsAbs,
		PolicyFileAbs:    policyAbs,
		PolicyID:         policyID,
		Symbol:           CanonicalICPolicySymbol,
	}
	execSteps := []pipeline.CommandSpec{
		{Name: "validate_policy", Command: icArgv, CWD: workspace, Timeout: 3 * time.Minute},
		{Name: "basedpyright", Command: []string{uvBin, "run", "basedpyright"}, CWD: workspace, Timeout: 6 * time.Minute},
		{Name: "ruff_check", Command: []string{uvBin, "run", "ruff", "check"}, CWD: workspace, Timeout: 3 * time.Minute},
		{Name: "pytest", Command: []string{uvBin, "run", "pytest"}, CWD: workspace, Timeout: 12 * time.Minute},
	}
	runner := execpipeline.Runner{
		CommandRunner: execpipeline.ExecCommandRunner{},
		Allowlist:     allow,
	}
	execResults := runner.Run(ctx, execSteps)
	results = append(results, execResults...)

	classification := pipeline.Classify(results)
	_ = writePipelineArtifacts(validationDir, results, classification)
	if classification.HasFailures() {
		return failureFromClassification(results, &classification, proposal)
	}
	return pureoptimizer.ProposalValidationResult{
		Results:        results,
		Classification: &classification,
	}, nil
}
