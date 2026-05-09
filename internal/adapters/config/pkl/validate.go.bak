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
	ErrMissingBaselineSystemID             = errors.New("config: baseline system id is required")
	ErrMissingIterativeContextSystemID     = errors.New("config: iterative context system id is required")
	ErrMissingInterfaceID                  = errors.New("config: interface id is required")
	ErrMissingAgentModelProvider           = errors.New("config: agent model provider is required")
	ErrMissingAgentModelName               = errors.New("config: agent model name is required")
	ErrMissingScoringObjectivePath         = errors.New("config: scoring objective path is required")
	ErrEvaluationAgentMismatch             = errors.New("config: evaluation.agent must reference agents.evaluator")
	ErrEvaluationBaselineSystemMismatch    = errors.New("config: evaluation.baseline.system must reference systems.baseline")
	ErrEvaluationCandidateSystemMismatch   = errors.New("config: evaluation.candidate.system must reference systems.iterativeContext")
	ErrMissingSelectionPolicyArtifact      = errors.New("config: candidate selection policy artifact is required")
	ErrSelectionPolicyArtifactMismatch     = errors.New("config: evaluation.candidate.uses.selectionPolicy must reference artifacts.candidatePolicyRound001")
	ErrSelectionPolicyInterfaceMismatch    = errors.New("config: candidate selection policy must implement interfaces.iterativeContextSelectionPolicyV1")
	ErrPolicyArtifactPathRequired          = errors.New("config: policy artifact path is required")
	ErrPolicyArtifactPathMustBeRelative    = errors.New("config: policy artifact path must be relative")
	ErrPolicyProposalArtifactNameRequired  = errors.New("config: policy proposal artifact name is required")
	ErrPolicyProposalArtifactNameInvalid   = errors.New("config: policy proposal artifact name must be relative and must not contain '..'")
	ErrCompletedBundleArtifactPathRequired = errors.New("config: completed evaluation bundle path is required")
	ErrOptimizerAgentMismatch              = errors.New("config: optimization.agent must reference agents.optimizer")
	ErrOptimizationParentBundleMismatch    = errors.New("config: optimization.parentRun.bundle must reference artifacts.parentEvaluationRound001")
	ErrOptimizationTargetInputMismatch     = errors.New("config: optimization.target.input must reference artifacts.candidatePolicyRound001")
	ErrOptimizationTargetOutputMismatch    = errors.New("config: optimization.target.output must reference artifacts.candidatePolicyRound002")
	ErrOptimizationEvidenceSourceMismatch  = errors.New("config: optimization.evidence.from must reference artifacts.parentEvaluationRound001")
	ErrToolAllowEntryEmpty                 = errors.New("config: tool allow entries must be non-empty")
	ErrToolDenyEntryEmpty                  = errors.New("config: tool deny entries must be non-empty")
	ErrToolAllowDuplicate                  = errors.New("config: tool allow entries must not duplicate")
	ErrToolDenyDuplicate                   = errors.New("config: tool deny entries must not duplicate")
	ErrToolPolicyOverlap                   = errors.New("config: tool allow and deny entries must not overlap")
	ErrSystemPromptTooLarge                = errors.New("config: systemPrompt must be at most 8 KiB")
)

// Validate applies SearchBench-specific config checks after Pkl has resolved
// defaults and basic type constraints.
func Validate(experiment Experiment) error {
	if err := validateMode(experiment); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateDataset(experiment.Dataset); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateInterfaces(experiment.Interfaces); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateSystems(experiment.Systems); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateArtifacts(experiment.Artifacts); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateAgents(experiment.Agents); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateEvaluation(experiment); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateOptimization(experiment); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	return nil
}

func validateMode(experiment Experiment) error {
	switch experiment.Mode {
	case ModeEvaluation:
		if experiment.Evaluation == nil {
			return ErrMissingEvaluation
		}
		if experiment.Agents.Evaluator == nil {
			return ErrMissingEvaluator
		}
	case ModeOptimization:
		if experiment.Optimization == nil {
			return ErrMissingOptimization
		}
		if experiment.Agents.Optimizer == nil {
			return ErrMissingOptimizer
		}
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedMode, experiment.Mode)
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
	if strings.TrimSpace(systems.Baseline.Id) == "" {
		return ErrMissingBaselineSystemID
	}
	if strings.TrimSpace(systems.IterativeContext.Id) == "" {
		return ErrMissingIterativeContextSystemID
	}
	return nil
}

func validateArtifacts(artifacts Artifacts) error {
	if artifacts.CandidatePolicyRound001 != nil {
		if err := validatePolicyArtifact(*artifacts.CandidatePolicyRound001); err != nil {
			return err
		}
	}
	if artifacts.CandidatePolicyRound002 != nil {
		if err := validatePolicyProposalArtifact(*artifacts.CandidatePolicyRound002); err != nil {
			return err
		}
	}
	if artifacts.ParentEvaluationRound001 != nil {
		if err := validateCompletedBundleArtifact(*artifacts.ParentEvaluationRound001); err != nil {
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

func validateCompletedBundleArtifact(artifact CompletedEvaluationBundleArtifact) error {
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

func validateEvaluation(experiment Experiment) error {
	if experiment.Evaluation == nil {
		return nil
	}
	evaluation := *experiment.Evaluation

	if experiment.Agents.Evaluator == nil {
		return ErrMissingEvaluator
	}
	if !reflect.DeepEqual(*experiment.Agents.Evaluator, evaluation.Agent) {
		return ErrEvaluationAgentMismatch
	}
	if !reflect.DeepEqual(experiment.Systems.Baseline, evaluation.Baseline.System) {
		return ErrEvaluationBaselineSystemMismatch
	}
	if !reflect.DeepEqual(experiment.Systems.IterativeContext, evaluation.Candidate.System) {
		return ErrEvaluationCandidateSystemMismatch
	}
	if experiment.Artifacts.CandidatePolicyRound001 == nil {
		return ErrMissingSelectionPolicyArtifact
	}
	if !reflect.DeepEqual(*experiment.Artifacts.CandidatePolicyRound001, evaluation.Candidate.Uses.SelectionPolicy) {
		return ErrSelectionPolicyArtifactMismatch
	}
	if evaluation.Candidate.Uses.SelectionPolicy.Implements.Id != experiment.Interfaces.IterativeContextSelectionPolicyV1.Id {
		return ErrSelectionPolicyInterfaceMismatch
	}
	if strings.TrimSpace(evaluation.Scoring.Objective) == "" {
		return ErrMissingScoringObjectivePath
	}
	return nil
}

func validateOptimization(experiment Experiment) error {
	if experiment.Optimization == nil {
		return nil
	}
	optimization := *experiment.Optimization

	if experiment.Agents.Optimizer == nil {
		return ErrMissingOptimizer
	}
	if !reflect.DeepEqual(*experiment.Agents.Optimizer, optimization.Agent) {
		return ErrOptimizerAgentMismatch
	}
	if experiment.Artifacts.ParentEvaluationRound001 == nil || !reflect.DeepEqual(*experiment.Artifacts.ParentEvaluationRound001, optimization.ParentRun.Bundle) {
		return ErrOptimizationParentBundleMismatch
	}
	if experiment.Artifacts.CandidatePolicyRound001 == nil || !reflect.DeepEqual(*experiment.Artifacts.CandidatePolicyRound001, optimization.Target.Input) {
		return ErrOptimizationTargetInputMismatch
	}
	if experiment.Artifacts.CandidatePolicyRound002 == nil || !reflect.DeepEqual(*experiment.Artifacts.CandidatePolicyRound002, optimization.Target.Output) {
		return ErrOptimizationTargetOutputMismatch
	}
	if !reflect.DeepEqual(optimization.ParentRun.Bundle, optimization.Evidence.From) {
		return ErrOptimizationEvidenceSourceMismatch
	}
	return nil
}

func containsParentPath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(path))
	return clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../")
}
