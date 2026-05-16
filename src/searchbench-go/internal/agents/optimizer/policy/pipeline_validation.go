package policy

import (
	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// PipelineValidationFromPortSteps maps ports/pipeline step results into pure optimizer records.
func PipelineValidationFromPortSteps(results []pipeline.StepResult) pureoptimizer.PipelineValidationResult {
	steps := make([]pureoptimizer.PipelineStepResult, 0, len(results))
	ok := true
	for _, r := range results {
		if r.Failed() {
			ok = false
		}
		steps = append(steps, pureoptimizer.PipelineStepResult{
			Name:     r.Name,
			Passed:   r.Passed,
			CWD:      r.CWD,
			Command:  r.CommandString(),
			ExitCode: r.ExitCode,
		})
	}
	return pureoptimizer.PipelineValidationResult{OK: ok, Steps: steps}
}
