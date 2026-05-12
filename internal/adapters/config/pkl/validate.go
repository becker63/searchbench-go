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
	ErrValidationFailed                           = errors.New("config: validation failed")
	ErrUnsupportedMode                            = errors.New("config: unsupported mode")
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
	ErrPolicyArtifactPathMustBeRelative           = errors.New("config: policy artifact path must be relative")
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
	if err := validateMode(spec); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateDataset(spec.Dataset); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateInterfaces(spec.Interfaces); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validatePolicies(spec.Policies); err != nil {
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

func validatePolicies(policies Policies) error {
	if strings.TrimSpace(policies.Incumbent.Id) == "" {
		return ErrMissingIncumbentPolicyID
	}
	if strings.TrimSpace(policies.Challenger.Id) == "" {
		return ErrMissingChallengerPolicyID
	}
	return nil
}

func validateArtifacts(artifacts Artifacts) error {
	if artifacts.ChallengerPolicy != nil {
		if err := validatePolicyArtifact(*artifacts.ChallengerPolicy); err != nil {
			return err
		}
	}
	if artifacts.NextChallenger != nil {
		if err := validateNextChallengerArtifact(*artifacts.NextChallenger); err != nil {
			return err
		}
	}
	if artifacts.ParentRoundBundle != nil {
		if err := validateCompletedBundleArtifact(*artifacts.ParentRoundBundle); err != nil {
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

func validateNextChallengerArtifact(artifact NextChallengerArtifact) error {
	name := strings.TrimSpace(artifact.ArtifactName)
	if name == "" {
		return ErrNextChallengerArtifactNameRequired
	}
	if filepath.IsAbs(name) || containsParentPath(name) {
		return ErrNextChallengerArtifactNameInvalid
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
		if err := validateToolPolicy("agents.evaluator.tools", agents.Evaluator.Tools); err != nil {
			return err
		}
		if err := validateSystemPrompt("agents.evaluator.systemPrompt", agents.Evaluator.SystemPrompt); err != nil {
			return err
		}
	}
	if agents.Optimizer != nil {
		if err := validateAgentModel(agents.Optimizer.Model); err != nil {
			return err
		}
		// Structural shape only: the optimizer agent has no tool registry enforcement yet.
		if err := validateToolPolicy("agents.optimizer.tools", agents.Optimizer.Tools); err != nil {
			return err
		}
		if err := validateSystemPrompt("agents.optimizer.systemPrompt", agents.Optimizer.SystemPrompt); err != nil {
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
	if !reflect.DeepEqual(spec.Policies.Incumbent, evaluation.Incumbent.System) {
		return ErrEvaluationIncumbentPolicyMismatch
	}
	if !reflect.DeepEqual(spec.Policies.Challenger, evaluation.Challenger.System) {
		return ErrEvaluationChallengerPolicyMismatch
	}
	if spec.Artifacts.ChallengerPolicy == nil {
		return ErrMissingChallengerSelectionPolicyArtifact
	}
	if !reflect.DeepEqual(*spec.Artifacts.ChallengerPolicy, evaluation.Challenger.Uses.SelectionPolicy) {
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
	if spec.Artifacts.ParentRoundBundle == nil || !reflect.DeepEqual(*spec.Artifacts.ParentRoundBundle, optimization.ParentRound.Bundle) {
		return ErrOptimizationParentRoundBundleMismatch
	}
	if spec.Artifacts.ChallengerPolicy == nil || !reflect.DeepEqual(*spec.Artifacts.ChallengerPolicy, optimization.Target.Input) {
		return ErrOptimizationChallengerPolicyInputMismatch
	}
	if spec.Artifacts.NextChallenger == nil || !reflect.DeepEqual(*spec.Artifacts.NextChallenger, optimization.Target.Output) {
		return ErrOptimizationNextChallengerOutputMismatch
	}
	if !reflect.DeepEqual(optimization.ParentRound.Bundle, optimization.Evidence.From) {
		return ErrNextChallengerEvidenceSourceMismatch
	}
	return nil
}

func containsParentPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../")
}
