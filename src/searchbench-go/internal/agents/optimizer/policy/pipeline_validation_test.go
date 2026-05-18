package policy

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/ports/pipeline"
)

func TestPipelineValidationFromPortSteps(t *testing.T) {
	t.Parallel()
	results := []pipeline.StepResult{
		{Name: "stage_policy", Passed: true, CWD: "/ws", Command: []string{}},
		{Name: "pytest", Passed: false, CWD: "/ws", Command: []string{"uv", "run", "pytest"}, ExitCode: 1},
	}
	out := PipelineValidationFromPortSteps(results)
	if out.OK {
		t.Fatal("expected validation not ok")
	}
	if len(out.Steps) != 2 || out.Steps[1].ExitCode != 1 {
		t.Fatalf("steps: %+v", out.Steps)
	}
}
