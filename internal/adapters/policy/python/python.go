// Package python is the policy validation adapter that compiles a staged
// Python policy proposal via the system `python3` interpreter.
//
// It owns the os/exec subprocess and temporary directory effects so the
// optimizer flow stays effect-free at its boundary.
package python

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	executoreino "github.com/becker63/searchbench-go/internal/adapters/executor/eino"
	execpipeline "github.com/becker63/searchbench-go/internal/adapters/pipeline/exec"
	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Validate stages the proposal to a temporary directory and runs
// `python3 -m py_compile` to confirm it parses. It is the default
// optimizer.ValidateProposal implementation for Python policies.
func Validate(ctx context.Context, proposal pureoptimizer.NextChallengerProposal) (executoreino.ProposalValidationResult, *pureoptimizer.Failure) {
	stageDir, err := os.MkdirTemp("", "searchbench-optimizer-stage-*")
	if err != nil {
		return executoreino.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "create staged policy directory",
			Cause:   err,
		}
	}
	defer os.RemoveAll(stageDir)

	stagePath := filepath.Join(stageDir, proposal.ArtifactName)
	if err := os.WriteFile(stagePath, []byte(proposal.Code), 0o644); err != nil {
		return executoreino.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "write staged policy proposal",
			Cause:   err,
		}
	}

	pythonBin, err := resolvePython()
	if err != nil {
		step := pipeline.StepResult{
			Name:                "python_compile",
			Command:             []string{"python3", "-m", "py_compile", proposal.ArtifactName},
			CWD:                 stageDir,
			ExitCode:            -1,
			InfrastructureError: err,
		}
		classification := pipeline.Classify([]pipeline.StepResult{step})
		return executoreino.ProposalValidationResult{
				Results:        []pipeline.StepResult{step},
				Classification: &classification,
			}, &pureoptimizer.Failure{
				Phase:            pureoptimizer.PhaseRunPolicyPipeline,
				Kind:             pureoptimizer.FailureKindPolicyPipelineInfrastructure,
				Message:          "resolve python interpreter",
				Cause:            err,
				Retryable:        false,
				PipelineCategory: "infrastructure",
				PipelineFeedback: pipeline.FormatPipelineFeedback(classification, 1200),
			}
	}

	spec := pipeline.CommandSpec{
		Name:    "python_compile",
		Command: []string{pythonBin, "-m", "py_compile", proposal.ArtifactName},
		CWD:     stageDir,
		Timeout: 5 * time.Second,
	}
	result := execpipeline.ExecCommandRunner{}.Run(ctx, spec)
	result.Name = spec.Name
	result.Command = append([]string(nil), spec.Command...)
	result.CWD = spec.CWD
	if result.InfrastructureError == nil && result.ExitCode == 0 {
		result.Passed = true
	}
	results := []pipeline.StepResult{result}
	classification := pipeline.Classify(results)
	if classification.HasFailures() {
		kind := pureoptimizer.FailureKindPolicyPipelineFailed
		retryable := true
		category := "validation"
		if len(classification.InfrastructureFailures) > 0 {
			kind = pureoptimizer.FailureKindPolicyPipelineInfrastructure
			retryable = false
			category = "infrastructure"
		}
		return executoreino.ProposalValidationResult{
				Results:        results,
				Classification: &classification,
			}, &pureoptimizer.Failure{
				Phase:            pureoptimizer.PhaseRunPolicyPipeline,
				Kind:             kind,
				Message:          "policy validation pipeline failed",
				Retryable:        retryable,
				PipelineCategory: category,
				PipelineFeedback: pipeline.FormatPipelineFeedback(classification, 1200),
			}
	}

	return executoreino.ProposalValidationResult{
		Results:        results,
		Classification: &classification,
	}, nil
}

func resolvePython() (string, error) {
	for _, candidate := range []string{"python3", "python"} {
		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, nil
		}
	}
	return "", errors.New("python interpreter not found")
}
