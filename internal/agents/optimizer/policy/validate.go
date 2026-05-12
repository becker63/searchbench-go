package policy

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	execpipeline "github.com/becker63/searchbench-go/internal/adapters/pipeline/exec"
	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Validate stages the proposal and validates it. Iterative Context selection policies
// run the full SearchBench IC pipeline; other Python artifacts keep the lightweight
// py_compile precheck.
func Validate(ctx context.Context, proposal pureoptimizer.NextChallengerProposal) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	if proposal.InterfaceID == IterativeContextSelectionPolicyInterfaceID {
		return validateIterativeContextProposal(ctx, proposal)
	}
	return validatePythonCompileOnly(ctx, proposal)
}

func validatePythonCompileOnly(ctx context.Context, proposal pureoptimizer.NextChallengerProposal) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	stageDir, err := os.MkdirTemp("", "searchbench-optimizer-stage-*")
	if err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
			Phase:   pureoptimizer.PhaseWriteNextChallengerStage,
			Kind:    pureoptimizer.FailureKindPolicyStageWriteFailed,
			Message: "create staged policy directory",
			Cause:   err,
		}
	}
	defer func() { _ = os.RemoveAll(stageDir) }()

	stagePath := filepath.Join(stageDir, proposal.ArtifactName)
	if err := os.WriteFile(stagePath, []byte(proposal.Code), 0o644); err != nil {
		return pureoptimizer.ProposalValidationResult{}, &pureoptimizer.Failure{
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
		return pureoptimizer.ProposalValidationResult{
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
		return pureoptimizer.ProposalValidationResult{
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

	return pureoptimizer.ProposalValidationResult{
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
