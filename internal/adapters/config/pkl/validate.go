package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

const maxSystemPromptBytes = 8 * 1024

var (
	ErrValidationFailed                    = errors.New("config: validation failed")
	ErrUnsupportedMode                     = errors.New("config: unsupported mode")
	ErrMissingEvaluation                   = errors.New("config: evaluation config is required")
	ErrMissingOptimization                 = errors.New("config: optimization config is required")
	ErrMissingEvaluator                    = errors.New("config: agents.evaluator is required")
	ErrMissingOptimizer                    = errors.New("config: agents.optimizer is required")
	ErrMissingDatasetConfig                = errors.New("config: dataset config is required")
	ErrMissingDatasetSplit                 = errors.New("config: dataset split is required")
	ErrMissingIncumbentSystemID    = errors.New("config: incumbent system id is required")
	ErrMissingChallengerSystemID   = errors.New("config: challenger system id is required")
	ErrMissingInterfaceID                  = errors.New("config: interface id is required")
	ErrMissingAgentModelProvider           = errors.New("config: agent model provider is required")
	ErrMissingAgentModelName               = errors.New("config: agent model name is required")
	ErrMissingScoringObjectivePath         = errors.New("config: scoring objective path is required")
	ErrEvaluationAgentMismatch             = errors.New("config: evaluation.agent must reference agents.evaluator")
	ErrEvaluationIncumbentSystemMismatch  = errors.New("config: evaluation.incumbent.system must reference systems.incumbent")
	ErrEvaluationChallengerSystemMismatch   = errors.New("config: evaluation.challenger.system must reference systems.challenger")
	ErrMissingChallengerSelectionPolicyArtifact   = errors.New("config: challenger selection policy artifact is required")
	ErrChallengerSelectionPolicyArtifactMismatch  = errors.New("config: evaluation.challenger.uses.selectionPolicy must reference artifacts.challengerPolicyRound001")
	ErrChallengerSelectionPolicyInterfaceMismatch = errors.New("config: challenger selection policy must implement interfaces.iterativeContextSelectionPolicyV1")
	ErrPolicyArtifactPathRequired          = errors.New("config: policy artifact path is required")
	ErrPolicyArtifactPathMustBeRelative    = errors.New("config: policy artifact path must be relative")
	ErrPolicyProposalArtifactNameRequired  = errors.New("config: policy proposal artifact name is required")
	ErrPolicyProposalArtifactNameInvalid   = errors.New("config: policy proposal artifact name must be relative and must not contain '..'")
	ErrCompletedBundleArtifactPathRequired = errors.New("config: completed round bundle path is required")
	ErrOptimizerAgentMismatch              = errors.New("config: optimization.agent must reference agents.optimizer")
	ErrOptimizationParentRoundBundleMismatch       = errors.New("config: optimization.parentRound.bundle must reference artifacts.parentRound001Bundle")
	ErrOptimizationChallengerPolicyInputMismatch   = errors.New("config: optimization.target.input must reference artifacts.challengerPolicyRound001")
	ErrOptimizationNextChallengerOutputMismatch    = errors.New("config: optimization.target.output must reference artifacts.nextChallengerRound002")
	ErrOptimizationEvidenceSourceMismatch          = errors.New("config: optimization.evidence.from must reference artifacts.parentRound001Bundle")
	ErrToolAllowEntryEmpty                 = errors.New("config: tool allow entries must be non-empty")
	ErrToolDenyEntryEmpty                  = errors.New("config: tool deny entries must be non-empty")
	ErrToolAllowDuplicate                  = errors.New("config: tool allow entries must not duplicate")
	ErrToolDenyDuplicate                   = errors.New("config: tool deny entries must not duplicate")
	ErrToolPolicyOverlap                   = errors.New("config: tool allow and deny entries must not overlap")
	ErrSystemPromptTooLarge                = errors.New("config: systemPrompt must be at most 8 KiB")
)

// Validate applies SearchBench-specific config checks after Pkl has resolved
// defaults and basic type constraints.
func Validate(spec RoundSpec) error {
	if err := validateMode(spec); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateDataset(spec.Dataset); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateInterfaces(spec.Interfaces); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateSystems(spec.Systems); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateArtifacts(spec.Artifacts); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateAgents(spec.Agents); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateEvaluation(spec); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateOptimization(spec); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	return nil
}

func validateMode(spec RoundSpec) error {
	switch spec.Mode {
	case ModeEvaluation:
		if spec.Evaluation == nil {
			return ErrMissingEvaluation
		}
		if spec.Agents.Evaluator == nil {
			return ErrMissingEvaluator
		}
	case ModeOptimization:
		if spec.Optimization == nil {
			return ErrMissingOptimization
		}
		if spec.Agents.Optimizer == nil {
			return ErrMissingOptimizer
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedMode, spec.Mode)
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

func validateSystems(systems Systems) error {
	if strings.TrimSpace(systems.Incumbent.Id) == "" {
		return ErrMissingIncumbentSystemID
	}
	if strings.TrimSpace(systems.Challenger.Id) == "" {
		return ErrMissingChallengerSystemID
	}
	return nil
}

func validateArtifacts(artifacts Artifacts) error {
	if artifacts.ChallengerPolicyRound001 != nil {
		if err := validatePolicyArtifact(*artifacts.ChallengerPolicyRound001); err != nil {
			return err
		}
	}
	if artifacts.NextChallengerRound002 != nil {
		if err := validatePolicyProposalArtifact(*artifacts.NextChallengerRound002); err != nil {
			return err
		}
	}
	if artifacts.ParentRound001Bundle != nil {
		if err := validateCompletedBundleArtifact(*artifacts.ParentRound001Bundle); err != nil {
			return err
		}
	}
	return nil
}

func validatePolicyArtifact(artifact PolicyArtifact) error {
	path := strings.TrimSpace(artifact.Path)
	if path == "" {
		return ErrPolicyArtifactPathRequired
	}
	if filepath.IsAbs(path) {
		return ErrPolicyArtifactPathMustBeRelative
	}
	return nil
}

func validatePolicyProposalArtifact(artifact PolicyProposalArtifact) error {
	name := strings.TrimSpace(artifact.ArtifactName)
	if name == "" {
		return ErrPolicyProposalArtifactNameRequired
	}
	if filepath.IsAbs(name) || containsParentPath(name) {
		return ErrPolicyProposalArtifactNameInvalid
	}
	return nil
}

func validateCompletedBundleArtifact(artifact CompletedRoundBundleArtifact) error {
	if strings.TrimSpace(artifact.Path) == "" {
		return ErrCompletedBundleArtifactPathRequired
	}
	return nil
}

func validateAgents(agents Agents) error {
	if agents.Evaluator != nil {
		if err := validateAgentModel(agents.Evaluator.Model); err != nil {
			return err
		}
		if err := validateToolPolicy(agents.Evaluator.Tools); err != nil {
			return err
		}
		if err := validateSystemPrompt(agents.Evaluator.SystemPrompt); err != nil {
			return err
		}
	}
	if agents.Optimizer != nil {
		if err := validateAgentModel(agents.Optimizer.Model); err != nil {
			return err
		}
		if err := validateToolPolicy(agents.Optimizer.Tools); err != nil {
			return err
		}
		if err := validateSystemPrompt(agents.Optimizer.SystemPrompt); err != nil {
			return err
		}
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

func validateToolPolicy(policy AgentToolPolicy) error {
	seenAllow := make(map[string]struct{}, len(policy.Allow))
	for _, entry := range policy.Allow {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			return ErrToolAllowEntryEmpty
		}
		if _, ok := seenAllow[trimmed]; ok {
			return ErrToolAllowDuplicate
		}
		seenAllow[trimmed] = struct{}{}
	}

	seenDeny := make(map[string]struct{}, len(policy.Deny))
	for _, entry := range policy.Deny {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			return ErrToolDenyEntryEmpty
		}
		if _, ok := seenDeny[trimmed]; ok {
			return ErrToolDenyDuplicate
		}
		if _, ok := seenAllow[trimmed]; ok {
			return ErrToolPolicyOverlap
		}
		seenDeny[trimmed] = struct{}{}
	}

	return nil
}

func validateSystemPrompt(prompt *string) error {
	if prompt == nil {
		return nil
	}
	if len(*prompt) > maxSystemPromptBytes {
		return ErrSystemPromptTooLarge
	}
	return nil
}

func validateEvaluation(spec RoundSpec) error {
	if spec.Evaluation == nil {
		return nil
	}
	evaluation := *spec.Evaluation

	if spec.Agents.Evaluator == nil {
		return ErrMissingEvaluator
	}
	if !reflect.DeepEqual(*spec.Agents.Evaluator, evaluation.Agent) {
		return ErrEvaluationAgentMismatch
	}
	if !reflect.DeepEqual(spec.Systems.Incumbent, evaluation.Incumbent.System) {
		return ErrEvaluationIncumbentSystemMismatch
	}
	if !reflect.DeepEqual(spec.Systems.Challenger, evaluation.Challenger.System) {
		return ErrEvaluationChallengerSystemMismatch
	}
	if spec.Artifacts.ChallengerPolicyRound001 == nil {
		return ErrMissingChallengerSelectionPolicyArtifact
	}
	if !reflect.DeepEqual(*spec.Artifacts.ChallengerPolicyRound001, evaluation.Challenger.Uses.SelectionPolicy) {
		return ErrChallengerSelectionPolicyArtifactMismatch
	}
	if evaluation.Challenger.Uses.SelectionPolicy.Implements.Id != spec.Interfaces.IterativeContextSelectionPolicyV1.Id {
		return ErrChallengerSelectionPolicyInterfaceMismatch
	}
	if strings.TrimSpace(evaluation.Scoring.Objective) == "" {
		return ErrMissingScoringObjectivePath
	}
	return nil
}

func validateOptimization(spec RoundSpec) error {
	if spec.Optimization == nil {
		return nil
	}
	optimization := *spec.Optimization

	if spec.Agents.Optimizer == nil {
		return ErrMissingOptimizer
	}
	if !reflect.DeepEqual(*spec.Agents.Optimizer, optimization.Agent) {
		return ErrOptimizerAgentMismatch
	}
	if spec.Artifacts.ParentRound001Bundle == nil || !reflect.DeepEqual(*spec.Artifacts.ParentRound001Bundle, optimization.ParentRound.Bundle) {
		return ErrOptimizationParentRoundBundleMismatch
	}
	if spec.Artifacts.ChallengerPolicyRound001 == nil || !reflect.DeepEqual(*spec.Artifacts.ChallengerPolicyRound001, optimization.Target.Input) {
		return ErrOptimizationChallengerPolicyInputMismatch
	}
	if spec.Artifacts.NextChallengerRound002 == nil || !reflect.DeepEqual(*spec.Artifacts.NextChallengerRound002, optimization.Target.Output) {
		return ErrOptimizationNextChallengerOutputMismatch
	}
	if !reflect.DeepEqual(optimization.ParentRound.Bundle, optimization.Evidence.From) {
		return ErrOptimizationEvidenceSourceMismatch
	}
	return nil
}

func containsParentPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../")
}
