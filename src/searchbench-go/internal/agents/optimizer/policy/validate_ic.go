package policy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func validateIterativeContextProposal(ctx context.Context, proposal pureoptimizer.NextChallengerProposal) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	return ValidateProposalWithBuckDescriptor(ctx, proposal)
}

func writePipelineArtifacts(dir string, results []pipeline.StepResult, classification pipeline.Classification) error {
	simplified := make([]map[string]any, 0, len(results))
	for _, r := range results {
		row := map[string]any{
			"name":        r.Name,
			"passed":      r.Passed,
			"exit_code":   r.ExitCode,
			"timed_out":   r.TimedOut,
			"skipped":     r.Skipped,
			"duration_ms": r.Duration.Milliseconds(),
		}
		if r.InfrastructureError != nil {
			row["infrastructure_error"] = r.InfrastructureError.Error()
		}
		simplified = append(simplified, row)
	}
	payload := map[string]any{
		"steps":        simplified,
		"has_failures": classification.HasFailures(),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "pipeline-result.json"), data, 0o644)
}

func pipelineFailureResult(step pipeline.StepResult) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	results := []pipeline.StepResult{step}
	classification := pipeline.Classify(results)
	return failureFromClassification(results, &classification, pureoptimizer.NextChallengerProposal{})
}

func stagePolicyFailure(message string) pipeline.StepResult {
	return pipeline.StepResult{
		Name:     "stage_policy",
		Command:  []string{},
		ExitCode: 1,
		Stderr:   message,
	}
}

func infraFailure(proposal pureoptimizer.NextChallengerProposal, cause error) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	step := pipeline.StepResult{
		Name:                "policy_workspace",
		Command:             []string{},
		ExitCode:            -1,
		InfrastructureError: cause,
	}
	results := []pipeline.StepResult{step}
	classification := pipeline.Classify(results)
	res, fail := failureFromClassification(results, &classification, proposal)
	fail.Retryable = false
	fail.Kind = pureoptimizer.FailureKindPolicyPipelineInfrastructure
	return res, fail
}

func failureFromClassification(results []pipeline.StepResult, classification *pipeline.Classification, proposal pureoptimizer.NextChallengerProposal) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
	kind := pureoptimizer.FailureKindPolicyPipelineFailed
	retryable := true
	category := "validation"
	if len(classification.InfrastructureFailures) > 0 {
		kind = pureoptimizer.FailureKindPolicyPipelineInfrastructure
		retryable = false
		category = "infrastructure"
	}
	msg := "policy validation pipeline failed"
	if proposal.ArtifactID.String() != "" {
		msg = fmt.Sprintf("policy validation pipeline failed for %s", proposal.ArtifactID.String())
	}
	return pureoptimizer.ProposalValidationResult{
			Results:        results,
			Classification: classification,
		}, &pureoptimizer.Failure{
			Phase:            pureoptimizer.PhaseRunPolicyPipeline,
			Kind:             kind,
			Message:          msg,
			Retryable:        retryable,
			PipelineCategory: category,
			PipelineFeedback: pipeline.FormatPipelineFeedback(*classification, 2400),
			Cause:            errors.New(msg),
		}
}
