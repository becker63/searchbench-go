package eino

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/pipeline"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestPipelineFailurePreventsModelExecution(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		},
	)
	commandRunner := &scriptedCommandRunner{
		results: []pipeline.StepResult{
			{Name: "templ_generate", ExitCode: 0},
			{Name: "gofmt_check", ExitCode: 0},
			{
				Name:     "go_test",
				ExitCode: 1,
				Stderr:   "FAIL ./internal/executor/eino",
			},
		},
	}

	evaluator, err := New(Config{
		Model:         model,
		CommandRunner: commandRunner,
		PipelineSteps: []pipeline.CommandSpec{
			{Name: "templ_generate", Command: []string{"templ", "generate"}, CWD: "/repo"},
			{Name: "gofmt_check", Command: []string{"gofmt", "-l", "."}, CWD: "/repo"},
			{Name: "go_test", Command: []string{"go", "test", "./..."}, CWD: "/repo"},
		},
		WorkDir: "/repo",
		Now:     fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure == nil {
		t.Fatal("expected pipeline failure")
	}
	if got, want := result.Failure.Kind, FailureKindPipelineFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := result.Failure.Phase, PhaseClassifyPipeline; got != want {
		t.Fatalf("Failure.Phase = %q, want %q", got, want)
	}
	if result.Failure.PipelineClassification == nil {
		t.Fatal("PipelineClassification = nil, want classification")
	}
	if len(result.Failure.PipelineClassification.TestFailures) != 1 {
		t.Fatalf("len(TestFailures) = %d, want 1", len(result.Failure.PipelineClassification.TestFailures))
	}
	if !strings.Contains(result.Failure.PipelineFeedback, "## TEST FAILURES") {
		t.Fatalf("PipelineFeedback missing test-failure section:\n%s", result.Failure.PipelineFeedback)
	}
	if calls := model.Calls(); len(calls) != 0 {
		t.Fatalf("len(model.Calls()) = %d, want 0", len(calls))
	}
	if len(result.Attempts) != 0 {
		t.Fatalf("len(Attempts) = %d, want 0", len(result.Attempts))
	}
	if result.Failure.Recoverable {
		t.Fatal("Failure.Recoverable = true, want false")
	}
}

func TestSuccessfulPipelineAllowsModelExecution(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		},
	)

	evaluator, err := New(Config{
		Model:         model,
		CommandRunner: newPassingCommandRunner(),
		WorkDir:       "/repo",
		Now:           fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if calls := model.Calls(); len(calls) != 1 {
		t.Fatalf("len(model.Calls()) = %d, want 1", len(calls))
	}
	if len(result.PipelineResults) != 2 {
		t.Fatalf("len(PipelineResults) = %d, want 2", len(result.PipelineResults))
	}
}

func TestPipelineFailureRemainsNonRetryableWithExplicitRetryPolicy(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		},
	)
	commandRunner := &scriptedCommandRunner{
		results: []pipeline.StepResult{
			{
				Name:     "templ_generate",
				ExitCode: 1,
				Stderr:   "templ failed",
			},
		},
	}

	evaluator, err := New(Config{
		Model:         model,
		CommandRunner: commandRunner,
		PipelineSteps: []pipeline.CommandSpec{
			{Name: "templ_generate", Command: []string{"templ", "generate"}, CWD: "/repo"},
		},
		WorkDir: "/repo",
		RetryPolicy: &RetryPolicy{
			MaxAttempts:                3,
			RetryOnModelError:          true,
			RetryOnToolFailure:         true,
			RetryOnFinalizationFailure: true,
			RetryOnInvalidPrediction:   true,
		},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure == nil {
		t.Fatal("expected pipeline failure")
	}
	if got, want := result.Failure.Kind, FailureKindPipelineFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if result.Failure.Recoverable {
		t.Fatal("Failure.Recoverable = true, want false")
	}
	if got, want := len(model.Calls()), 0; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
	if got, want := len(result.Attempts), 0; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
}

func TestDefaultPipelineUsesOnlyLocalPreflightSteps(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		},
	)

	commandRunner := newPassingCommandRunner()
	evaluator, err := New(Config{
		Model:         model,
		CommandRunner: commandRunner,
		WorkDir:       "/repo",
		Now:           fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if got, want := len(result.PipelineResults), 2; got != want {
		t.Fatalf("len(PipelineResults) = %d, want %d", got, want)
	}
	if got, want := len(commandRunner.Calls()), 2; got != want {
		t.Fatalf("len(commandRunner.Calls()) = %d, want %d", got, want)
	}
	if got, want := commandRunner.Calls()[1].Name, "gofmt_check"; got != want {
		t.Fatalf("second step name = %q, want %q", got, want)
	}
}
