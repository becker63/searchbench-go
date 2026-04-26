package eino

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestEvaluatorConstructionIsCold(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel()
	evaluator, err := New(Config{Model: model})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if evaluator == nil {
		t.Fatal("expected evaluator")
	}
	if calls := model.Calls(); len(calls) != 0 {
		t.Fatalf("len(model.Calls()) = %d, want 0", len(calls))
	}
}

func TestEvaluatorRunSuccessWithToolCall(t *testing.T) {
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
			Message: schema.AssistantMessage(`{"predicted_files":["./SRC/Main.go","src\\main.go"],"reasoning":"fake tool matched the retry path"}`, nil),
		},
	)

	evaluator, err := New(Config{
		Model: model,
		Tools: []tool.BaseTool{fakeTool{
			name:   "fake_resolve",
			result: `{"resolved_path":"src/main.go"}`,
		}},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	spec := sampleRunSpec()
	result := evaluator.Run(context.Background(), spec)
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if result.Executed == nil {
		t.Fatal("expected executed run")
	}

	if got, want := result.Phases, []Phase{
		PhaseRenderPrompt,
		PhaseRunEvaluator,
		PhaseFinalizePrediction,
		PhaseComplete,
	}; !equalPhases(got, want) {
		t.Fatalf("Phases = %#v, want %#v", got, want)
	}

	if got, want := result.Executed.Prediction.Files, []domain.RepoRelPath{"src/main.go"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("Prediction.Files = %#v, want %#v", got, want)
	}
	if got, want := result.Executed.Prediction.Reasoning, "fake tool matched the retry path"; got != want {
		t.Fatalf("Prediction.Reasoning = %q, want %q", got, want)
	}

	calls := model.Calls()
	if len(calls) != 2 {
		t.Fatalf("len(model.Calls()) = %d, want 2", len(calls))
	}

	secondCallMessages := calls[1].Messages
	if len(secondCallMessages) < 3 {
		t.Fatalf("len(secondCallMessages) = %d, want at least 3", len(secondCallMessages))
	}
	toolMessage := secondCallMessages[len(secondCallMessages)-1]
	if toolMessage.Role != schema.Tool {
		t.Fatalf("toolMessage.Role = %q, want %q", toolMessage.Role, schema.Tool)
	}
	if !strings.Contains(toolMessage.Content, "src/main.go") {
		t.Fatalf("tool message content = %q, want tool result", toolMessage.Content)
	}
}

func TestEvaluatorRunFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		model         *modeltest.ScriptedModel
		tools         []tool.BaseTool
		renderPrompt  RenderPromptFunc
		wantKind      FailureKind
		wantPhase     Phase
		wantMessageIn string
	}{
		{
			name: "prompt render failure",
			model: modeltest.NewScriptedModel(
				modeltest.ScriptedResponse{
					Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
				},
			),
			renderPrompt: func(context.Context, domain.TaskSpec, []string) (string, error) {
				return "", errors.New("render blew up")
			},
			wantKind:      FailureKindPromptRenderFailed,
			wantPhase:     PhaseRenderPrompt,
			wantMessageIn: "render blew up",
		},
		{
			name: "model error",
			model: modeltest.NewScriptedModel(
				modeltest.ScriptedResponse{Err: errors.New("model exploded")},
			),
			wantKind:      FailureKindEvaluatorFailed,
			wantPhase:     PhaseRunEvaluator,
			wantMessageIn: "model exploded",
		},
		{
			name: "malformed final prediction",
			model: modeltest.NewScriptedModel(
				modeltest.ScriptedResponse{
					Message: schema.AssistantMessage(`not json`, nil),
				},
			),
			wantKind:      FailureKindFinalizationFailed,
			wantPhase:     PhaseFinalizePrediction,
			wantMessageIn: "parse final prediction JSON",
		},
		{
			name: "empty prediction",
			model: modeltest.NewScriptedModel(
				modeltest.ScriptedResponse{
					Message: schema.AssistantMessage(`{"predicted_files":[],"reasoning":"none"}`, nil),
				},
			),
			wantKind:      FailureKindInvalidPrediction,
			wantPhase:     PhaseFinalizePrediction,
			wantMessageIn: "predicted files are required",
		},
		{
			name: "tool failure",
			model: modeltest.NewScriptedModel(
				modeltest.ScriptedResponse{
					Message: schema.AssistantMessage("", []schema.ToolCall{{
						ID: "call-1",
						Function: schema.FunctionCall{
							Name:      "fake_resolve",
							Arguments: `{"query":"retry interceptor"}`,
						},
					}}),
				},
			),
			tools: []tool.BaseTool{fakeTool{
				name: "fake_resolve",
				err:  errors.New("fixture tool failed"),
			}},
			wantKind:      FailureKindToolCallFailed,
			wantPhase:     PhaseRunEvaluator,
			wantMessageIn: "fixture tool failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			evaluator, err := New(Config{
				Model:        tt.model,
				Tools:        tt.tools,
				RenderPrompt: tt.renderPrompt,
				Now:          fixedClock(),
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			result := evaluator.Run(context.Background(), sampleRunSpec())
			if result.Failure == nil {
				t.Fatal("expected failure")
			}
			if got, want := result.Failure.Kind, tt.wantKind; got != want {
				t.Fatalf("Failure.Kind = %q, want %q", got, want)
			}
			if got, want := result.Failure.Phase, tt.wantPhase; got != want {
				t.Fatalf("Failure.Phase = %q, want %q", got, want)
			}
			if !strings.Contains(result.Failure.Error(), tt.wantMessageIn) {
				t.Fatalf("Failure.Error() = %q, want substring %q", result.Failure.Error(), tt.wantMessageIn)
			}
		})
	}
}

func TestEvaluatorExecuteReturnsFailureAsError(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":[]}`, nil),
		},
	)

	evaluator, err := New(Config{Model: model, Now: fixedClock()})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = evaluator.Execute(context.Background(), sampleRunSpec())
	if err == nil {
		t.Fatal("expected error")
	}
	var failure *Failure
	if !errors.As(err, &failure) {
		t.Fatalf("expected *Failure, got %T", err)
	}
	if got, want := failure.Kind, FailureKindInvalidPrediction; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
}

func TestEvaluatorUsesExistingRunModelOnSuccess(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"files":["src/main.go"],"reasoning":"fixture"}`, nil),
		},
	)

	evaluator, err := New(Config{Model: model, Now: fixedClock()})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	executed, err := evaluator.Execute(context.Background(), sampleRunSpec())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got, want := executed.Run.Phase().Name(), run.PhaseExecuted; got != want {
		t.Fatalf("executed phase = %q, want %q", got, want)
	}
	if got, want := executed.Prediction.Reasoning, "fixture"; got != want {
		t.Fatalf("Prediction.Reasoning = %q, want %q", got, want)
	}
}

type fakeTool struct {
	name   string
	result string
	err    error
}

func (f fakeTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: f.name,
		Desc: "Resolve a deterministic fake repository path.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Desc:     "The symbol or issue hint to resolve.",
				Required: true,
				Type:     schema.String,
			},
		}),
	}, nil
}

func (f fakeTool) InvokableRun(_ context.Context, _ string, _ ...tool.Option) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.result, nil
}

var _ tool.InvokableTool = fakeTool{}

func sampleRunSpec() run.Spec {
	task := domain.TaskSpec{
		ID:        domain.TaskID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("square/okhttp"),
			SHA:  domain.RepoSHA("abc123"),
		},
		Input: domain.TaskInput{
			Title: "Crash when retrying HTTP request",
			Body:  "The client crashes when the retry interceptor replays a request.",
		},
		Oracle: domain.TaskOracle{
			GoldFiles: []domain.RepoRelPath{"internal/should/not/leak.go"},
		},
	}

	system := domain.SystemSpec{
		ID:      domain.SystemID("eino-minimal"),
		Name:    "Minimal Eino Evaluator",
		Backend: domain.BackendFake,
		Model: domain.ModelSpec{
			Provider: "fixture",
			Name:     "scripted",
		},
		PromptBundle: domain.PromptBundleRef{Name: "evaluator"},
	}

	return run.NewSpec(domain.RunID("run-1"), task, system)
}

func fixedClock() func() time.Time {
	fixed := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return fixed }
}

func equalPhases(got []Phase, want []Phase) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func ExampleFinalizePrediction() {
	prediction, _, _ := FinalizePrediction(`{"predicted_files":["./SRC/Main.go"],"reasoning":"fixture"}`)
	fmt.Println(prediction.Files[0])
	// Output: src/main.go
}
