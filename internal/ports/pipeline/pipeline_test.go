package pipeline

import (
	"errors"
	"strings"
	"testing"
)

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
