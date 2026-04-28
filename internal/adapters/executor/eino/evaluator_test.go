package eino

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	cloudcallbacks "github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	evaluatorcallbacks "github.com/becker63/searchbench-go/internal/adapters/executor/eino/callbacks"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	evaluatorprompt "github.com/becker63/searchbench-go/internal/pure/prompts/evaluator"
	"github.com/becker63/searchbench-go/internal/pure/run"
	"github.com/becker63/searchbench-go/internal/pure/usage"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestEvaluatorConstructionIsCold(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel()
	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
	})
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
		Model:   model,
		WorkDir: "/repo",
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
		PhasePrepareCallbacks,
		PhasePrepareUsageAccounting,
		PhaseRunEvaluator,
		PhaseFinalizeUsage,
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
	if got, want := result.UsageSummary.CallCount, 2; got != want {
		t.Fatalf("UsageSummary.CallCount = %d, want %d", got, want)
	}
	if len(result.UsageRecords) != 2 {
		t.Fatalf("len(UsageRecords) = %d, want 2", len(result.UsageRecords))
	}
	if result.Executed.Usage.Empty() {
		t.Fatal("Executed.Usage = empty, want usage summary")
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

func TestEvaluatorUsageAccountingWorksWithoutTracing(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"],"reasoning":"usage without tracing"}`, nil),
		},
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
		t.Fatal("expected executed run")
	}
	if got, want := result.UsageSummary.CallCount, 1; got != want {
		t.Fatalf("UsageSummary.CallCount = %d, want %d", got, want)
	}
	if len(result.UsageRecords) != 1 {
		t.Fatalf("len(UsageRecords) = %d, want 1", len(result.UsageRecords))
	}
	if got := result.UsageRecords[0].Source; got != usage.SourceEstimated {
		t.Fatalf("UsageRecords[0].Source = %q, want %q", got, usage.SourceEstimated)
	}
}

func TestEvaluatorRunCanDisableUsageAccounting(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"],"reasoning":"usage disabled"}`, nil),
		},
	)

	evaluator, err := New(Config{
		Model:                  model,
		WorkDir:                "/repo",
		DisableUsageAccounting: true,
		UsageCollectorFactory: func(run.Spec) (*usage.Collector, error) {
			return nil, errors.New("usage collector should not be constructed when accounting is disabled")
		},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if result.Executed == nil {
		t.Fatal("expected executed run")
	}
	if result.UsageSummary.CallCount != 0 {
		t.Fatalf("UsageSummary.CallCount = %d, want 0", result.UsageSummary.CallCount)
	}
	if len(result.UsageRecords) != 0 {
		t.Fatalf("len(UsageRecords) = %d, want 0", len(result.UsageRecords))
	}
	if !result.Executed.Usage.Empty() {
		t.Fatalf("Executed.Usage = %#v, want empty usage when disabled", result.Executed.Usage)
	}
}

func TestEvaluatorRunCanAttachCallbacks(t *testing.T) {
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
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"],"reasoning":"callback observed execution"}`, nil),
		},
	)

	recorder := &evaluatorcallbacks.FakeTestRecorder{}
	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Tools: []tool.BaseTool{fakeTool{
			name:   "fake_resolve",
			result: `{"resolved_path":"src/main.go"}`,
		}},
		CallbackFactories: []evaluatorcallbacks.Factory{
			evaluatorcallbacks.NewFakeTestCallbackFactory(recorder),
		},
		Now: fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if result.Executed == nil {
		t.Fatal("expected executed run")
	}

	// The fake callback is only proving the seam here: construction happened
	// during prepare_callbacks, then Eino invoked the attached handler while the
	// agent/model/tool execution ran.
	snapshot := recorder.Snapshot()
	if got, want := snapshot.Constructed, 1; got != want {
		t.Fatalf("Constructed = %d, want %d", got, want)
	}
	if snapshot.Attached == 0 || snapshot.AgentStarts == 0 || snapshot.AgentEnds == 0 {
		t.Fatalf("agent callback snapshot = %#v, want agent lifecycle events", snapshot)
	}
	if snapshot.ModelStarts == 0 || snapshot.ModelEnds == 0 {
		t.Fatalf("model callback snapshot = %#v, want model lifecycle events", snapshot)
	}
	if snapshot.ToolStarts == 0 || snapshot.ToolEnds == 0 {
		t.Fatalf("tool callback snapshot = %#v, want tool lifecycle events", snapshot)
	}
	if got, want := result.UsageSummary.CallCount, 2; got != want {
		t.Fatalf("UsageSummary.CallCount = %d, want %d", got, want)
	}
}

func TestEvaluatorCallbackSetupFailureReturnsTypedFailure(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		CallbackFactories: []evaluatorcallbacks.Factory{
			func(context.Context) (cloudcallbacks.Handler, error) {
				return nil, errors.New("fixture callback setup failed")
			},
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

	// Callback setup is its own harness phase. If the callback factory fails, the
	// evaluator must fail closed before handing control to Eino execution.
	if got, want := result.Failure.Kind, FailureKindCallbackSetupFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := result.Failure.Phase, PhasePrepareCallbacks; got != want {
		t.Fatalf("Failure.Phase = %q, want %q", got, want)
	}

	// Cold setup means no model call is allowed to happen after callback
	// construction fails.
	if got, want := len(model.Calls()), 0; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
	if !strings.Contains(result.Failure.Error(), "fixture callback setup failed") {
		t.Fatalf("Failure.Error() = %q, want callback setup detail", result.Failure.Error())
	}
}

func TestEvaluatorUsageAccountingSetupFailureReturnsTypedFailure(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		UsageCollectorFactory: func(run.Spec) (*usage.Collector, error) {
			return nil, errors.New("fixture usage setup failed")
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
	// Usage-accounting setup is separate from generic callback construction so
	// the harness can fail closed before evaluation if the run-local collector
	// cannot be created or attached.
	if got, want := result.Failure.Kind, FailureKindUsageAccountingSetupFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := result.Failure.Phase, PhasePrepareUsageAccounting; got != want {
		t.Fatalf("Failure.Phase = %q, want %q", got, want)
	}
	if got, want := len(model.Calls()), 0; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
}

func TestEvaluatorRunCanUseMultipleToolCallsAndModelTurns(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("", []schema.ToolCall{{
				ID: "call-1",
				Function: schema.FunctionCall{
					Name:      "fake_search",
					Arguments: `{"query":"retry interceptor"}`,
				},
			}}),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("", []schema.ToolCall{{
				ID: "call-2",
				Function: schema.FunctionCall{
					Name:      "fake_read",
					Arguments: `{"path":"src/main.go"}`,
				},
			}}),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"],"reasoning":"combined search and file read evidence"}`, nil),
		},
	)

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Tools: []tool.BaseTool{
			fakeTool{name: "fake_search", result: `{"hits":["src/main.go"]}`},
			fakeTool{name: "fake_read", result: `{"path":"src/main.go","snippet":"func retry() {}"}`},
		},
		RetryPolicy: &RetryPolicy{MaxAttempts: 1},
		Now:         fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), sampleRunSpec())
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if result.Executed == nil {
		t.Fatal("expected executed run")
	}
	if got, want := len(result.Attempts), 1; got != want {
		t.Fatalf("len(Attempts) = %d, want %d", got, want)
	}
	if got, want := len(model.Calls()), 3; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
	if got, want := result.Executed.Prediction.Files, []domain.RepoRelPath{"src/main.go"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("Prediction.Files = %#v, want %#v", got, want)
	}
	if got, want := result.Executed.Prediction.Reasoning, "combined search and file read evidence"; got != want {
		t.Fatalf("Prediction.Reasoning = %q, want %q", got, want)
	}

	secondCallMessages := model.Calls()[1].Messages
	lastSecondCallMessage := secondCallMessages[len(secondCallMessages)-1]
	if lastSecondCallMessage.Role != schema.Tool || !strings.Contains(lastSecondCallMessage.Content, "src/main.go") {
		t.Fatalf("second call last message = %#v, want tool result for fake_search", lastSecondCallMessage)
	}

	thirdCallMessages := model.Calls()[2].Messages
	lastThirdCallMessage := thirdCallMessages[len(thirdCallMessages)-1]
	if lastThirdCallMessage.Role != schema.Tool || !strings.Contains(lastThirdCallMessage.Content, "func retry() {}") {
		t.Fatalf("third call last message = %#v, want tool result for fake_read", lastThirdCallMessage)
	}
}

func TestEvaluatorRunUsesSystemRuntimeMaxStepsBound(t *testing.T) {
	t.Parallel()

	model := modeltest.NewScriptedModel(
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("", []schema.ToolCall{{
				ID: "call-1",
				Function: schema.FunctionCall{
					Name:      "fake_search",
					Arguments: `{"query":"retry interceptor"}`,
				},
			}}),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("", []schema.ToolCall{{
				ID: "call-2",
				Function: schema.FunctionCall{
					Name:      "fake_read",
					Arguments: `{"path":"src/main.go"}`,
				},
			}}),
		},
		modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"],"reasoning":"would have succeeded on the second turn"}`, nil),
		},
	)

	spec := sampleRunSpec()
	spec.System.Runtime.MaxSteps = 1

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Tools: []tool.BaseTool{
			fakeTool{name: "fake_search", result: `{"hits":["src/main.go"]}`},
			fakeTool{name: "fake_read", result: `{"path":"src/main.go","snippet":"func retry() {}"}`},
		},
		RetryPolicy: &RetryPolicy{MaxAttempts: 1},
		Now:         fixedClock(),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	result := evaluator.Run(context.Background(), spec)
	if result.Failure == nil {
		t.Fatal("expected failure")
	}
	if got, want := result.Failure.Kind, FailureKindEvaluatorFailed; got != want {
		t.Fatalf("Failure.Kind = %q, want %q", got, want)
	}
	if got, want := len(model.Calls()), 1; got != want {
		t.Fatalf("len(model.Calls()) = %d, want %d", got, want)
	}
	if !strings.Contains(strings.ToLower(result.Failure.Error()), "iteration") {
		t.Fatalf("Failure.Error() = %q, want iteration-limit detail", result.Failure.Error())
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
			renderPrompt: func(context.Context, evaluatorprompt.Input) (string, error) {
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
				WorkDir:      "/repo",
				Tools:        tt.tools,
				RenderPrompt: tt.renderPrompt,
				RetryPolicy:  &RetryPolicy{MaxAttempts: 1},
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

	evaluator, err := New(Config{
		Model:       model,
		WorkDir:     "/repo",
		RetryPolicy: &RetryPolicy{MaxAttempts: 1},
		Now:         fixedClock(),
	})
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

	evaluator, err := New(Config{
		Model:   model,
		WorkDir: "/repo",
		Now:     fixedClock(),
	})
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
