package execpipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	ports "github.com/becker63/searchbench-go/internal/ports/pipeline"
)

func TestRunnerRecordsTypedStepResult(t *testing.T) {
	t.Parallel()

	runner := &scriptedRunner{
		results: []ports.StepResult{{
			ExitCode: 1,
			Stdout:   "out",
			Stderr:   "err",
			Duration: 25 * time.Millisecond,
		}},
	}

	results := Runner{
		CommandRunner: runner,
		Allowlist:     DefaultAllowlist(),
	}.Run(context.Background(), []ports.CommandSpec{{
		Name:    "go_test",
		Command: []string{"go", "test", "./..."},
		CWD:     "/repo",
	}})

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	got := results[0]
	if got.Name != "go_test" {
		t.Fatalf("Name = %q, want go_test", got.Name)
	}
	if got.CommandString() != "go test ./..." {
		t.Fatalf("CommandString() = %q, want %q", got.CommandString(), "go test ./...")
	}
	if got.CWD != "/repo" {
		t.Fatalf("CWD = %q, want /repo", got.CWD)
	}
	if got.Passed {
		t.Fatal("Passed = true, want false")
	}
	if got.ExitCode != 1 {
		t.Fatalf("ExitCode = %d, want 1", got.ExitCode)
	}
	if got.Stdout != "out" {
		t.Fatalf("Stdout = %q, want out", got.Stdout)
	}
	if got.Stderr != "err" {
		t.Fatalf("Stderr = %q, want err", got.Stderr)
	}
	if got.Duration != 25*time.Millisecond {
		t.Fatalf("Duration = %v, want 25ms", got.Duration)
	}
}

func TestRunnerRejectsDisallowedCommandWithoutExecuting(t *testing.T) {
	t.Parallel()

	runner := &scriptedRunner{
		results: []ports.StepResult{{ExitCode: 0}},
	}

	results := Runner{
		CommandRunner: runner,
		Allowlist:     DefaultAllowlist(),
	}.Run(context.Background(), []ports.CommandSpec{{
		Name:    "rm_repo",
		Command: []string{"rm", "-rf", "."},
		CWD:     "/repo",
	}})

	if len(runner.calls) != 0 {
		t.Fatalf("len(runner.calls) = %d, want 0", len(runner.calls))
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].InfrastructureError == nil {
		t.Fatal("InfrastructureError = nil, want disallowed-command failure")
	}
}

func TestExecCommandRunnerSetsDurationOnSuccess(t *testing.T) {
	t.Parallel()

	result := ExecCommandRunner{}.Run(context.Background(), ports.CommandSpec{
		Name:    "go_version",
		Command: []string{"go", "version"},
	})

	if result.Duration <= 0 {
		t.Fatalf("Duration = %v, want > 0", result.Duration)
	}
	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.InfrastructureError != nil {
		t.Fatalf("InfrastructureError = %v, want nil", result.InfrastructureError)
	}
}

func TestExecCommandRunnerSetsDurationOnInfrastructureFailure(t *testing.T) {
	t.Parallel()

	result := ExecCommandRunner{}.Run(context.Background(), ports.CommandSpec{
		Name:    "missing_command",
		Command: []string{"searchbench-definitely-missing-command"},
	})

	if result.Duration <= 0 {
		t.Fatalf("Duration = %v, want > 0", result.Duration)
	}
	if result.InfrastructureError == nil {
		t.Fatal("InfrastructureError = nil, want non-nil")
	}
	if result.Passed {
		t.Fatal("Passed = true, want false")
	}
}

type scriptedRunner struct {
	results []ports.StepResult
	calls   []ports.CommandSpec
}

func (r *scriptedRunner) Run(_ context.Context, spec ports.CommandSpec) ports.StepResult {
	r.calls = append(r.calls, spec)
	if len(r.results) == 0 {
		return ports.StepResult{
			ExitCode:            -1,
			InfrastructureError: errors.New("no scripted results remaining"),
		}
	}

	result := r.results[0]
	r.results = r.results[1:]
	return result
}
