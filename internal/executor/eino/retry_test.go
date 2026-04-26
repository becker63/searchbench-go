package eino

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	evaluatorprompt "github.com/becker63/searchbench-go/internal/prompts/evaluator"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestMalformedFinalOutputRetriesAndSucceeds(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage("not json", nil)},
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":["./SRC/Main.go"],"reasoning":"retry succeeded"}`, nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Now:     fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if result.Executed == nil {
		t.Fatal("Executed = nil, want success")
	}
	if got, want := len(result.Attempts), 2; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
	if result.Attempts[0].Failure == nil || result.Attempts[0].Failure.Kind != FailureKindFinalizationFailed {
		t.Fatalf("first attempt failure = %#v, want finalization_failed", result.Attempts[0].Failure)
	}
	if got, want := len(model.Calls()), 2; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
	secondPrompt := model.Calls()[1].Messages[0].Content
	if !strings.Contains(secondPrompt, "Previous attempt returned malformed JSON.") {
		t.Fatalf("second prompt missing retry feedback:\n%s", secondPrompt)
	}
	if strings.Contains(secondPrompt, "internal/should/not/leak.go") {
		t.Fatalf("second prompt leaked oracle data:\n%s", secondPrompt)
	}
}

func TestEmptyPredictionRetriesAccordingToPolicy(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":[]}`, nil)},
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Now:     fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if result.Executed == nil {
		t.Fatal("Executed = nil, want success")
	}
	if got, want := len(result.Attempts), 2; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
	if result.Attempts[0].Failure == nil || result.Attempts[0].Failure.Kind != FailureKindInvalidPrediction {
		t.Fatalf("first attempt failure = %#v, want invalid_prediction", result.Attempts[0].Failure)
	}
}

func TestRetriesExhaustedReturnsTypedFailure(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage("not json", nil)},
		modeltest.ScriptedResponse{Message: schema.AssistantMessage("still not json", nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Now:     fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure == nil {
		t.Fatal("expected failure")
	}
	if got, want := result.Failure.Kind, FailureKindRetriesExhausted; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := result.Failure.Phase, PhaseExhausted; got != want {
		t.Fatalf("Failure.Phase = %q, want %q", got, want)
	}
	if got, want := len(result.Attempts), 2; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
}

func TestPromptRenderFailureIsNotRetried(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		RenderPrompt: func(context.Context, evaluatorprompt.Input) (string, error) {
			return "", errors.New("render prompt failed")
		},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure == nil {
		t.Fatal("expected failure")
	}
	if got, want := result.Failure.Kind, FailureKindPromptRenderFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := len(result.Attempts), 1; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
	if calls := model.Calls(); len(calls) != 0 {
		t.Fatalf("len(model.Calls()) = %d, want 0", len(calls))
	}
}

func TestRecoverableToolFailureCanRetryWhenEnabled(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("", []schema.ToolCall{{
				ID: "call-1",
				Function: schema.FunctionCall{
					Name:      "fake_resolve",
					Arguments: `{"query":"retry interceptor"}`,
				},
			}}),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("", []schema.ToolCall{{
				ID: "call-2",
				Function: schema.FunctionCall{
					Name:      "fake_resolve",
					Arguments: `{"query":"retry interceptor"}`,
				},
			}}),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Tools: []tool.BaseTool{fakeTool{
			name: "fake_resolve",
			err:  recoverableToolErr{message: "temporary tool outage"},
		}},
		RetryPolicy: &RetryPolicy{
			MaxAttempts:                2,
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
		t.Fatal("expected failure after repeated tool error")
	}
	if got, want := result.Failure.Kind, FailureKindRetriesExhausted; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
}

type recoverableToolErr struct {
	message string
}

func (e recoverableToolErr) Error() string { return e.message }

func (e recoverableToolErr) Recoverable() bool { return true }

func TestNilRetryPolicyUsesDefaults(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage("not json", nil)},
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Now:     fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if got, want := len(result.Attempts), 2; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
}

func TestExplicitRetryPolicyCanDisableFinalizationRetry(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage("not json", nil)},
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		RetryPolicy: &RetryPolicy{
			MaxAttempts:                2,
			RetryOnModelError:          true,
			RetryOnToolFailure:         false,
			RetryOnFinalizationFailure: false,
			RetryOnInvalidPrediction:   true,
		},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure == nil {
		t.Fatal("expected failure")
	}
	if got, want := result.Failure.Kind, FailureKindFinalizationFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := len(result.Attempts), 1; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
	if got, want := len(model.Calls()), 1; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
}

func TestExplicitRetryPolicyCanDisableInvalidPredictionRetry(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":[]}`, nil)},
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		RetryPolicy: &RetryPolicy{
			MaxAttempts:                2,
			RetryOnModelError:          true,
			RetryOnToolFailure:         false,
			RetryOnFinalizationFailure: true,
			RetryOnInvalidPrediction:   false,
		},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure == nil {
		t.Fatal("expected failure")
	}
	if got, want := result.Failure.Kind, FailureKindInvalidPrediction; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := len(result.Attempts), 1; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
	if got, want := len(model.Calls()), 1; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
}

func TestExplicitRetryPolicyNormalizesInvalidMaxAttemptsToOne(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{Message: schema.AssistantMessage("not json", nil)},
		modeltest.ScriptedResponse{Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil)},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		RetryPolicy: &RetryPolicy{
			MaxAttempts:                0,
			RetryOnModelError:          true,
			RetryOnToolFailure:         false,
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
		t.Fatal("expected failure")
	}
	if got, want := result.Failure.Kind, FailureKindFinalizationFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := len(result.Attempts), 1; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
	if got, want := len(model.Calls()), 1; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
}
