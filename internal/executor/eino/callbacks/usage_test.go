package callbacks

import (
	"context"
	"testing"

	einocallbacks "github.com/cloudwego/eino/callbacks"
	einocomponents "github.com/cloudwego/eino/components"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/usage"
)

func TestUsageCallbackConstructionIsCold(t *testing.T) {
	t.Parallel()

	collector, err := usage.NewCollector(usage.Config{
		DefaultProvider: "fixture",
		DefaultModel:    "scripted",
	})
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	callback, err := NewUsageCallback(UsageConfig{
		Phase:           "run_evaluator",
		DefaultProvider: "fixture",
		DefaultModel:    "scripted",
	})
	if err != nil {
		t.Fatalf("NewUsageCallback() error = %v", err)
	}
	if err := callback.AttachCollector(collector); err != nil {
		t.Fatalf("AttachCollector() error = %v", err)
	}

	// Construction plus attachment is allowed to allocate local state, but it
	// must not record any model usage before Eino actually starts a model call.
	if got := len(collector.Records()); got != 0 {
		t.Fatalf("len(collector.Records()) = %d, want 0", got)
	}
}

func TestUsageCallbackProviderUsageCreatesReportedRecord(t *testing.T) {
	t.Parallel()

	collector, _, handler := newUsageCallbackFixture(t)
	info := &einocallbacks.RunInfo{Name: "model-node", Component: einocomponents.ComponentOfChatModel}

	ctx := handler.OnStart(context.Background(), info, &einomodel.CallbackInput{
		Messages: []*schema.Message{
			schema.UserMessage("find retry interceptor"),
		},
		Config: &einomodel.Config{Model: "scripted"},
	})
	handler.OnEnd(ctx, info, &einomodel.CallbackOutput{
		Message: schema.AssistantMessage(`{"predicted_files":["src/main.go"]}`, nil),
		Config:  &einomodel.Config{Model: "scripted"},
		TokenUsage: &einomodel.TokenUsage{
			PromptTokens:     11,
			CompletionTokens: 4,
			TotalTokens:      15,
		},
	})

	records := collector.Records()
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	record := records[0]
	if got, want := record.Source, usage.SourceReported; got != want {
		t.Fatalf("record.Source = %q, want %q", got, want)
	}
	if got, want := record.InputTokens, domain.TokenCount(11); got != want {
		t.Fatalf("record.InputTokens = %d, want %d", got, want)
	}
	if got, want := record.OutputTokens, domain.TokenCount(4); got != want {
		t.Fatalf("record.OutputTokens = %d, want %d", got, want)
	}
	if got, want := record.TotalTokens, domain.TokenCount(15); got != want {
		t.Fatalf("record.TotalTokens = %d, want %d", got, want)
	}
}

func TestUsageCallbackMissingProviderUsageCreatesEstimatedRecord(t *testing.T) {
	t.Parallel()

	collector, _, handler := newUsageCallbackFixture(t)
	info := &einocallbacks.RunInfo{Name: "model-node", Component: einocomponents.ComponentOfChatModel}

	ctx := handler.OnStart(context.Background(), info, &einomodel.CallbackInput{
		Messages: []*schema.Message{
			schema.UserMessage("find retry interceptor"),
		},
		Config: &einomodel.Config{Model: "scripted"},
	})
	handler.OnEnd(ctx, info, &einomodel.CallbackOutput{
		Message: schema.AssistantMessage("src/main.go", nil),
		Config:  &einomodel.Config{Model: "scripted"},
	})

	records := collector.Records()
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	record := records[0]
	if got, want := record.Source, usage.SourceEstimated; got != want {
		t.Fatalf("record.Source = %q, want %q", got, want)
	}
	if record.InputTokens == 0 || record.OutputTokens == 0 || record.TotalTokens == 0 {
		t.Fatalf("record = %#v, want estimated token counts", record)
	}
}

func TestUsageCallbackComposesWithPeerCallbacks(t *testing.T) {
	t.Parallel()

	collector, _, usageHandler := newUsageCallbackFixture(t)
	fakeRecorder := &FakeTestRecorder{}
	peerHandlers, err := BuildCallbacks(context.Background(), Config{
		Factories: []Factory{
			NewFakeTestCallbackFactory(fakeRecorder),
		},
	})
	if err != nil {
		t.Fatalf("BuildCallbacks() error = %v", err)
	}

	// The fake test callback stands in for future tracing/observability siblings:
	// the usage callback must behave like a composable peer rather than a special
	// callback that monopolizes the execution seam.
	handlers := append(peerHandlers, usageHandler)
	info := &einocallbacks.RunInfo{Name: "model-node", Component: einocomponents.ComponentOfChatModel}

	contexts := make([]context.Context, len(handlers))
	for i, handler := range handlers {
		contexts[i] = handler.OnStart(context.Background(), info, &einomodel.CallbackInput{
			Messages: []*schema.Message{schema.UserMessage("find retry interceptor")},
			Config:   &einomodel.Config{Model: "scripted"},
		})
	}
	for i, handler := range handlers {
		handler.OnEnd(contexts[i], info, &einomodel.CallbackOutput{
			Message: schema.AssistantMessage("src/main.go", nil),
			Config:  &einomodel.Config{Model: "scripted"},
		})
	}

	records := collector.Records()
	if got := len(records); got != 1 {
		t.Fatalf("len(collector.Records()) = %d, want 1", got)
	}
	if snapshot := fakeRecorder.Snapshot(); snapshot.ModelStarts == 0 || snapshot.ModelEnds == 0 {
		t.Fatalf("fake recorder snapshot = %#v, want peer callback events", snapshot)
	}
}

func newUsageCallbackFixture(t *testing.T) (*usage.Collector, *UsageCallback, einocallbacks.Handler) {
	t.Helper()

	collector, err := usage.NewCollector(usage.Config{
		DefaultProvider: "fixture",
		DefaultModel:    "scripted",
	})
	if err != nil {
		t.Fatalf("NewCollector() error = %v", err)
	}

	callback, err := NewUsageCallback(UsageConfig{
		Phase:           "run_evaluator",
		DefaultProvider: "fixture",
		DefaultModel:    "scripted",
	})
	if err != nil {
		t.Fatalf("NewUsageCallback() error = %v", err)
	}
	if err := callback.AttachCollector(collector); err != nil {
		t.Fatalf("AttachCollector() error = %v", err)
	}
	return collector, callback, callback.Handler()
}
