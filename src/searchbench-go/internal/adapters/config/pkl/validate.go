package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/backend"
)

const maxSystemPromptBytes = 8 * 1024

var (
	ErrValidationFailed                           = errors.New("config: validation failed")
	ErrUnsupportedMode                            = errors.New("config: unsupported mode")
	ErrMissingRoundManifest                       = errors.New("config: round block is required")
	ErrMissingRoundID                             = errors.New("config: round.id is required")
	ErrMissingRoundIncumbent                      = errors.New("config: round.incumbent is required for from-scratch rounds")
	ErrMissingRoundMatches                        = errors.New("config: round.matches is required for from-scratch rounds")
	ErrMissingRoundEvaluator                      = errors.New("config: round.evaluator is required for from-scratch rounds")
	ErrMissingRoundScoring                        = errors.New("config: round.scoring is required for from-scratch rounds")
	ErrMissingRoundChallenger                     = errors.New("config: round.challenger must define selectionPolicy or generate")
	ErrRoundChallengerConflict                    = errors.New("config: round.challenger must not define both selectionPolicy and generate")
	ErrRoundGenerateArtifactNameInvalid           = errors.New("config: round.challenger.generate.artifactName must be relative and must not contain '..'")
	ErrMissingEvaluation                          = errors.New("config: evaluation config is required")
	ErrMissingOptimization                        = errors.New("config: optimization config is required")
	ErrMissingEvaluator                           = errors.New("config: agents.evaluator is required")
	ErrMissingOptimizer                           = errors.New("config: agents.optimizer is required")
	ErrMissingDatasetConfig                       = errors.New("config: dataset config is required")
	ErrMissingDatasetSplit                        = errors.New("config: dataset split is required")
	ErrMissingIncumbentPolicyID                   = errors.New("config: incumbent policy id is required")
	ErrMissingChallengerPolicyID                  = errors.New("config: challenger policy id is required")
	ErrMissingInterfaceID                         = errors.New("config: interface id is required")
	ErrMissingAgentModelProvider                  = errors.New("config: agent model provider is required")
	ErrMissingAgentModelName                      = errors.New("config: agent model name is required")
	ErrMissingScoringObjectivePath                = errors.New("config: scoring objective path is required")
	ErrEvaluationAgentMismatch                    = errors.New("config: evaluation.agent must reference agents.evaluator")
	ErrEvaluationIncumbentPolicyMismatch          = errors.New("config: evaluation.incumbent.system must reference policies.incumbent")
	ErrEvaluationChallengerPolicyMismatch         = errors.New("config: evaluation.challenger.system must reference policies.challenger")
	ErrMissingChallengerSelectionPolicyArtifact   = errors.New("config: challenger selection policy artifact is required")
	ErrChallengerSelectionPolicyArtifactMismatch  = errors.New("config: evaluation.challenger.uses.selectionPolicy must reference artifacts.challengerPolicy")
	ErrChallengerSelectionPolicyInterfaceMismatch = errors.New("config: challenger selection policy must implement interfaces.iterativeContextSelectionPolicyV1")
	ErrPolicyArtifactPathRequired                 = errors.New("config: policy artifact path is required")
	ErrNextChallengerArtifactNameRequired         = errors.New("config: next challenger artifact name is required")
	ErrNextChallengerArtifactNameInvalid          = errors.New("config: next challenger artifact name must be relative and must not contain '..'")
	ErrCompletedBundleArtifactPathRequired        = errors.New("config: completed round bundle path is required")
	ErrOptimizerAgentMismatch                     = errors.New("config: optimization.agent must reference agents.optimizer")
	ErrOptimizationParentRoundBundleMismatch      = errors.New("config: optimization.parentRound.bundle must reference artifacts.parentRoundBundle")
	ErrOptimizationChallengerPolicyInputMismatch  = errors.New("config: optimization.target.input must reference artifacts.challengerPolicy")
	ErrOptimizationNextChallengerOutputMismatch   = errors.New("config: optimization.target.output must reference artifacts.nextChallenger")
	ErrNextChallengerEvidenceSourceMismatch       = errors.New("config: optimization.evidence.from must reference artifacts.parentRoundBundle")
	ErrToolAllowEntryEmpty                        = errors.New("config: tool allow entries must be non-empty")
	ErrToolDenyEntryEmpty                         = errors.New("config: tool deny entries must be non-empty")
	ErrToolAllowDuplicate                         = errors.New("config: tool allow entries must not duplicate")
	ErrToolDenyDuplicate                          = errors.New("config: tool deny entries must not duplicate")
	ErrToolPolicyOverlap                          = errors.New("config: tool allow and deny entries must not overlap")
	ErrSystemPromptTooLarge                       = errors.New("config: systemPrompt must be at most 8 KiB")
)

// Validate applies SearchBench-specific config checks after Pkl has resolved
// defaults and basic type constraints.
func Validate(spec RoundSpec) error {
	if spec.Round == nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, ErrMissingRoundManifest)
	}
	if err := validateInterfaces(spec.Interfaces); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateRoundManifest(spec); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	return nil
}

func validateRoundManifest(spec RoundSpec) error {
	round := spec.Round
	if round == nil {
		return nil
	}
	if strings.TrimSpace(round.Id) == "" {
		return ErrMissingRoundID
	}
	if round.Continues == nil {
		if round.Incumbent == nil {
			return ErrMissingRoundIncumbent
		}
		if round.Matches == nil {
			return ErrMissingRoundMatches
		}
		if round.Evaluator == nil {
			return ErrMissingRoundEvaluator
		}
		if round.Scoring == nil {
			return ErrMissingRoundScoring
		}
	}
	if round.Incumbent != nil {
		if err := validateRoundPolicy("round.incumbent", *round.Incumbent); err != nil {
			return err
		}
	}
	if err := validateRoundChallenger(round.Challenger); err != nil {
		return err
	}
	if round.Matches != nil {
		if err := validateDataset(*round.Matches); err != nil {
			return err
		}
	}
	if round.Evaluator != nil {
		if err := validateAgentModel(round.Evaluator.Model); err != nil {
			return err
		}
		if err := validateToolPolicy("round.evaluator.tools", round.Evaluator.Tools); err != nil {
			return err
		}
		if err := validateSystemPrompt("round.evaluator.systemPrompt", round.Evaluator.SystemPrompt); err != nil {
			return err
		}
	}
	if round.Scoring != nil && strings.TrimSpace(round.Scoring.Objective) == "" {
		return ErrMissingScoringObjectivePath
	}
	return nil
}

func validateRoundPolicy(fieldPath string, policy RoundPolicy) error {
	if strings.TrimSpace(policy.System.Id) == "" {
		return fmt.Errorf("%s.system.id: %w", fieldPath, ErrMissingIncumbentPolicyID)
	}
	if err := validateICWorkspaceSeed(fieldPath+".system", policy.System); err != nil {
		return err
	}
	if policy.SelectionPolicy != nil {
		if err := validatePolicyArtifact(*policy.SelectionPolicy); err != nil {
			return err
		}
	}
	return nil
}

func validateRoundChallenger(challenger RoundChallenger) error {
	if challenger.SelectionPolicy != nil && challenger.Generate != nil {
		return ErrRoundChallengerConflict
	}
	if challenger.SelectionPolicy == nil && challenger.Generate == nil {
		return ErrMissingRoundChallenger
	}
	if challenger.System != nil {
		if strings.TrimSpace(challenger.System.Id) == "" {
			return ErrMissingChallengerPolicyID
		}
		if err := validateICWorkspaceSeed("round.challenger.system", *challenger.System); err != nil {
			return err
		}
	}
	if challenger.SelectionPolicy != nil {
		if err := validatePolicyArtifact(*challenger.SelectionPolicy); err != nil {
			return err
		}
	}
	if challenger.Generate != nil {
		if err := validateAgentModel(challenger.Generate.Optimizer.Model); err != nil {
			return err
		}
		if err := validateToolPolicy("round.challenger.generate.optimizer.tools", challenger.Generate.Optimizer.Tools); err != nil {
			return err
		}
		if err := validateSystemPrompt("round.challenger.generate.optimizer.systemPrompt", challenger.Generate.Optimizer.SystemPrompt); err != nil {
			return err
		}
		name := strings.TrimSpace(challenger.Generate.ArtifactName)
		if filepath.IsAbs(name) || containsParentPath(name) || name == "" {
			return ErrRoundGenerateArtifactNameInvalid
		}
	}
	return nil
}

func validateICWorkspaceSeed(fieldPath string, sys System) error {
	if sys.Backend != backend.IterativeContext {
		return nil
	}
	if sys.Runtime.WorkspaceSeed == nil {
		return nil
	}
	ws := sys.Runtime.WorkspaceSeed
	if err := ValidateWorkspaceSeedConfig(string(ws.Provider), ws.LocalPath); err != nil {
		return fmt.Errorf("%s.runtime.workspaceSeed: %w", fieldPath, err)
	}
	return nil
}

func validateDataset(dataset Dataset) error {
	if strings.TrimSpace(dataset.Config) == "" {
		return ErrMissingDatasetConfig
	}
	if strings.TrimSpace(dataset.Split) == "" {
		return ErrMissingDatasetSplit
	}
	return nil
}

func validateInterfaces(interfaces Interfaces) error {
	if strings.TrimSpace(interfaces.IterativeContextSelectionPolicyV1.Id) == "" {
		return ErrMissingInterfaceID
	}
	return nil
}

func validatePolicyArtifact(artifact PolicyArtifact) error {
	path := strings.TrimSpace(artifact.Path)
	if path == "" {
		return ErrPolicyArtifactPathRequired
	}
	return nil
}

func validateAgentModel(model Model) error {
	if strings.TrimSpace(model.Provider.String()) == "" {
		return ErrMissingAgentModelProvider
	}
	if strings.TrimSpace(model.Name) == "" {
		return ErrMissingAgentModelName
	}
	return nil
}

func validateToolPolicy(prefix string, policy AgentToolPolicy) error {
	seenAllow := make(map[string]struct{}, len(policy.Allow))
	for _, entry := range policy.Allow {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			return fmt.Errorf("%s.allow contains empty tool name: %w", prefix, ErrToolAllowEntryEmpty)
		}
		if _, ok := seenAllow[trimmed]; ok {
			return fmt.Errorf("%s.allow contains duplicate tool %q: %w", prefix, trimmed, ErrToolAllowDuplicate)
		}
		seenAllow[trimmed] = struct{}{}
	}

	seenDeny := make(map[string]struct{}, len(policy.Deny))
	for _, entry := range policy.Deny {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			return fmt.Errorf("%s.deny contains empty tool name: %w", prefix, ErrToolDenyEntryEmpty)
		}
		if _, ok := seenDeny[trimmed]; ok {
			return fmt.Errorf("%s.deny contains duplicate tool %q: %w", prefix, trimmed, ErrToolDenyDuplicate)
		}
		if _, ok := seenAllow[trimmed]; ok {
			return fmt.Errorf("%s.deny overlaps %s.allow for tool %q: %w", prefix, prefix, trimmed, ErrToolPolicyOverlap)
		}
		seenDeny[trimmed] = struct{}{}
	}

	return nil
}

func validateSystemPrompt(fieldPath string, prompt *string) error {
	if prompt == nil {
		return nil
	}
	if len(*prompt) > maxSystemPromptBytes {
		return fmt.Errorf("%s exceeds max length %d bytes: %w", fieldPath, maxSystemPromptBytes, ErrSystemPromptTooLarge)
	}
	return nil
}

func containsParentPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../")
}
