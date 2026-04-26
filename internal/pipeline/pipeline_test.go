package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunnerRecordsTypedStepResult(t *testing.T) {
	t.Parallel()

	runner := &scriptedRunner{
		results: []StepResult{{
			ExitCode: 1,
			Stdout:   "out",
			Stderr:   "err",
			Duration: 25 * time.Millisecond,
		}},
	}

	results := Runner{
		CommandRunner: runner,
		Allowlist:     DefaultAllowlist(),
	}.Run(context.Background(), []CommandSpec{{
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

func TestClassifySuccessfulPipelineRecordsPassedSteps(t *testing.T) {
	t.Parallel()

	classification := Classify([]StepResult{
		{Name: "templ_generate", Command: []string{"templ", "generate"}, Passed: true},
		{Name: "gofmt_check", Command: []string{"gofmt", "-l", "."}, Passed: true},
	})

	if classification.HasFailures() {
		t.Fatal("HasFailures() = true, want false")
	}
	if len(classification.PassedSteps) != 2 {
		t.Fatalf("len(PassedSteps) = %d, want 2", len(classification.PassedSteps))
	}
}

func TestClassifyTemplFailureAsGenerationFailure(t *testing.T) {
	t.Parallel()

	classification := Classify([]StepResult{{
		Name:     "templ_generate",
		Command:  []string{"templ", "generate"},
		ExitCode: 1,
	}})

	if len(classification.GenerationFailures) != 1 {
		t.Fatalf("len(GenerationFailures) = %d, want 1", len(classification.GenerationFailures))
	}
}

func TestClassifyGofmtFailureAsFormatError(t *testing.T) {
	t.Parallel()

	classification := Classify([]StepResult{{
		Name:     "gofmt_check",
		Command:  []string{"gofmt", "-l", "."},
		ExitCode: 1,
	}})

	if len(classification.FormatErrors) != 1 {
		t.Fatalf("len(FormatErrors) = %d, want 1", len(classification.FormatErrors))
	}
}

func TestClassifyGoTestFailures(t *testing.T) {
	t.Parallel()

	typeErr := Classify([]StepResult{{
		Name:     "go_test",
		Command:  []string{"go", "test", "./..."},
		ExitCode: 1,
		Stderr:   "undefined: missingSymbol",
	}})
	if len(typeErr.TypeErrors) != 1 {
		t.Fatalf("len(TypeErrors) = %d, want 1", len(typeErr.TypeErrors))
	}

	testErr := Classify([]StepResult{{
		Name:     "go_test",
		Command:  []string{"go", "test", "./..."},
		ExitCode: 1,
		Stderr:   "--- FAIL: TestEvaluator (0.00s)",
	}})
	if len(testErr.TestFailures) != 1 {
		t.Fatalf("len(TestFailures) = %d, want 1", len(testErr.TestFailures))
	}
}

func TestClassifyInfrastructureFailures(t *testing.T) {
	t.Parallel()

	classification := Classify([]StepResult{{
		Name:                "go_test",
		Command:             []string{"go", "test", "./..."},
		InfrastructureError: errors.New("exec: \"go\": executable file not found"),
	}})

	if len(classification.InfrastructureFailures) != 1 {
		t.Fatalf("len(InfrastructureFailures) = %d, want 1", len(classification.InfrastructureFailures))
	}
}

func TestRunnerRejectsDisallowedCommandWithoutExecuting(t *testing.T) {
	t.Parallel()

	runner := &scriptedRunner{
		results: []StepResult{{ExitCode: 0}},
	}

	results := Runner{
		CommandRunner: runner,
		Allowlist:     DefaultAllowlist(),
	}.Run(context.Background(), []CommandSpec{{
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

func TestFormatPipelineFeedbackIsDeterministicAndBounded(t *testing.T) {
	t.Parallel()

	classification := Classification{
		GenerationFailures: []StepResult{{
			Name:     "templ_generate",
			Command:  []string{"templ", "generate"},
			ExitCode: 1,
			Stderr:   strings.Repeat("x", 60),
		}},
		PassedSteps: []StepResult{{
			Name:    "gofmt_check",
			Command: []string{"gofmt", "-l", "."},
			Passed:  true,
		}},
	}

	feedback := FormatPipelineFeedback(classification, 220)
	if !strings.Contains(feedback, "## GENERATION FAILURES") {
		t.Fatalf("feedback missing generation section:\n%s", feedback)
	}
	if !strings.Contains(feedback, "## PASSED STEPS") {
		t.Fatalf("feedback missing passed section:\n%s", feedback)
	}
	if len([]rune(feedback)) > 220 {
		t.Fatalf("feedback length = %d, want <= 220", len([]rune(feedback)))
	}
}

func TestExecCommandRunnerSetsDurationOnSuccess(t *testing.T) {
	t.Parallel()

	result := ExecCommandRunner{}.Run(context.Background(), CommandSpec{
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

	result := ExecCommandRunner{}.Run(context.Background(), CommandSpec{
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
	results []StepResult
	calls   []CommandSpec
}

func (r *scriptedRunner) Run(_ context.Context, spec CommandSpec) StepResult {
	r.calls = append(r.calls, spec)
	if len(r.results) == 0 {
		return StepResult{
			ExitCode:            -1,
			InfrastructureError: errors.New("no scripted results remaining"),
		}
	}

	result := r.results[0]
	r.results = r.results[1:]
	return result
}
