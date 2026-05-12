package round

import (
	"context"

	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// stubOptimizerPipelinePass returns a validator that skips heavy IC tooling while
// preserving the optimizer retry/plumbing shape for fast round tests.
func stubOptimizerPipelinePass() pureoptimizer.ValidateProposalFunc {
	return func(context.Context, pureoptimizer.NextChallengerProposal) (pureoptimizer.ProposalValidationResult, *pureoptimizer.Failure) {
		step := pipeline.StepResult{Name: "stub_validate", Passed: true, ExitCode: 0}
		classification := pipeline.Classify([]pipeline.StepResult{step})
		return pureoptimizer.ProposalValidationResult{
			Results:        []pipeline.StepResult{step},
			Classification: &classification,
		}, nil
	}
}
