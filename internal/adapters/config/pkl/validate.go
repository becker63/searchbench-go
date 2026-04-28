package config

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrValidationFailed            = errors.New("config: validation failed")
	ErrUnsupportedMode             = errors.New("config: unsupported mode")
	ErrWriterNotAllowed            = errors.New("config: writer is not allowed for evaluator_only mode")
	ErrWriterRequired              = errors.New("config: writer is required for optimization modes")
	ErrMissingDatasetConfig        = errors.New("config: dataset config is required")
	ErrMissingDatasetSplit         = errors.New("config: dataset split is required")
	ErrMissingBaselineSystemID     = errors.New("config: baseline system id is required")
	ErrMissingCandidateSystemID    = errors.New("config: candidate system id is required")
	ErrMissingEvaluatorProvider    = errors.New("config: evaluator model provider is required")
	ErrMissingEvaluatorModelName   = errors.New("config: evaluator model name is required")
	ErrMissingScoringObjectivePath = errors.New("config: scoring objective path is required")
	ErrEmptyPipelineName           = errors.New("config: writer pipeline name is required")
	ErrEmptyPipelineStepName       = errors.New("config: writer pipeline step name is required")
	ErrEmptyPipelineStepArgv       = errors.New("config: writer pipeline step argv is required")
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
	if err := validateSystems(experiment.Systems); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateEvaluator(experiment.Evaluator); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateWriter(experiment.Writer); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	if err := validateScoring(experiment.Scoring); err != nil {
		return fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}
	return nil
}

func validateMode(experiment Experiment) error {
	switch experiment.Mode {
	case ModeEvaluatorOnly:
		if experiment.Writer != nil && experiment.Writer.Enabled {
			return ErrWriterNotAllowed
		}
	case ModeWriterOptimization, ModeOptimizationKickoff:
		if experiment.Writer == nil || !experiment.Writer.Enabled {
			return ErrWriterRequired
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

func validateSystems(systems Systems) error {
	if strings.TrimSpace(systems.Baseline.Id) == "" {
		return ErrMissingBaselineSystemID
	}
	if strings.TrimSpace(systems.Candidate.Id) == "" {
		return ErrMissingCandidateSystemID
	}
	return nil
}

func validateEvaluator(evaluator Evaluator) error {
	if strings.TrimSpace(evaluator.Model.Provider.String()) == "" {
		return ErrMissingEvaluatorProvider
	}
	if strings.TrimSpace(evaluator.Model.Name) == "" {
		return ErrMissingEvaluatorModelName
	}
	return nil
}

func validateWriter(writer *Writer) error {
	if writer == nil {
		return nil
	}
	if writer.Pipeline == nil {
		return nil
	}
	if strings.TrimSpace(writer.Pipeline.Name) == "" {
		return ErrEmptyPipelineName
	}
	for _, step := range writer.Pipeline.Steps {
		if strings.TrimSpace(step.Name) == "" {
			return ErrEmptyPipelineStepName
		}
		if len(step.Argv) == 0 {
			return ErrEmptyPipelineStepArgv
		}
	}
	return nil
}

func validateScoring(scoring Scoring) error {
	if strings.TrimSpace(scoring.Objective) == "" {
		return ErrMissingScoringObjectivePath
	}
	return nil
}
