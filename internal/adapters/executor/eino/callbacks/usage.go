package callbacks

import (
	"context"
	"errors"
	"sync"

	einocallbacks "github.com/cloudwego/eino/callbacks"
	einocomponents "github.com/cloudwego/eino/components"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/pure/usage"
)

type usageCallKey struct{}

// UsageConfig configures the harness-owned usage accounting callback.
type UsageConfig struct {
	Phase           string
	DefaultProvider string
	DefaultModel    string
}

// UsageCallback is the concrete Eino callback used for harness-owned token
// accounting.
//
// It stays execution-layer only: Eino callbacks feed provider-neutral data into
// an attached usage collector, and the collector owns all normalization.
type UsageCallback struct {
	mu        sync.RWMutex
	collector *usage.Collector

	phase           string
	defaultProvider string
	defaultModel    string
	handler         einocallbacks.Handler
}

// NewUsageCallback constructs a cold usage callback without starting model
// execution.
func NewUsageCallback(config UsageConfig) (*UsageCallback, error) {
	if config.Phase == "" {
		return nil, errors.New("usage callback phase is required")
	}

	callback := &UsageCallback{
		phase:           config.Phase,
		defaultProvider: config.DefaultProvider,
		defaultModel:    config.DefaultModel,
	}
	callback.handler = einocallbacks.NewHandlerBuilder().
		OnStartFn(callback.onStart).
		OnEndFn(callback.onEnd).
		Build()
	return callback, nil
}

// AttachCollector connects the callback to a run-local usage collector.
func (c *UsageCallback) AttachCollector(collector *usage.Collector) error {
	if collector == nil {
		return errors.New("usage collector is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.collector = collector
	return nil
}

// Handler returns the Eino callback handler for this usage callback.
func (c *UsageCallback) Handler() einocallbacks.Handler {
	if c == nil {
		return nil
	}
	return c.handler
}

func (c *UsageCallback) onStart(ctx context.Context, info *einocallbacks.RunInfo, input einocallbacks.CallbackInput) context.Context {
	if info == nil || info.Component != einocomponents.ComponentOfChatModel {
		return ctx
	}

	collector := c.getCollector()
	if collector == nil {
		return ctx
	}

	modelInput := einomodel.ConvCallbackInput(input)
	if modelInput == nil {
		return ctx
	}

	callID := collector.StartCall(usage.StartEvent{
		Phase:    c.phase,
		Node:     info.Name,
		Provider: c.defaultProvider,
		Model:    chooseModelName(modelInput.Config, c.defaultModel),
		Input:    extractInputTexts(modelInput.Messages),
	})
	return context.WithValue(ctx, usageCallKey{}, callID)
}

func (c *UsageCallback) onEnd(ctx context.Context, info *einocallbacks.RunInfo, output einocallbacks.CallbackOutput) context.Context {
	if info == nil || info.Component != einocomponents.ComponentOfChatModel {
		return ctx
	}

	collector := c.getCollector()
	if collector == nil {
		return ctx
	}

	callID, _ := ctx.Value(usageCallKey{}).(int)
	modelOutput := einomodel.ConvCallbackOutput(output)
	if modelOutput == nil {
		return ctx
	}

	collector.EndCall(callID, usage.EndEvent{
		Provider: c.defaultProvider,
		Model:    chooseModelName(modelOutput.Config, c.defaultModel),
		Output:   extractOutputTexts(modelOutput.Message),
		Reported: normalizeReportedUsage(modelOutput.TokenUsage),
	})
	return ctx
}

func (c *UsageCallback) getCollector() *usage.Collector {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.collector
}

func chooseModelName(config *einomodel.Config, fallback string) string {
	if config != nil && config.Model != "" {
		return config.Model
	}
	return fallback
}

func normalizeReportedUsage(tokenUsage *einomodel.TokenUsage) usage.ReportedUsage {
	if tokenUsage == nil {
		return usage.ReportedUsage{}
	}

	reported := usage.ReportedUsage{
		InputTokens:  usage.ReportedCount(int64(tokenUsage.PromptTokens)),
		OutputTokens: usage.ReportedCount(int64(tokenUsage.CompletionTokens)),
		TotalTokens:  usage.ReportedCount(int64(tokenUsage.TotalTokens)),
	}
	if tokenUsage.PromptTokens == 0 {
		reported.InputTokens.Set = false
	}
	if tokenUsage.CompletionTokens == 0 {
		reported.OutputTokens.Set = false
	}
	if tokenUsage.TotalTokens == 0 {
		reported.TotalTokens.Set = false
	}
	return reported
}

func extractInputTexts(messages []*schema.Message) []string {
	if len(messages) == 0 {
		return nil
	}

	parts := make([]string, 0, len(messages)*2)
	for _, message := range messages {
		if message == nil {
			continue
		}
		if message.Role != "" {
			parts = append(parts, string(message.Role))
		}
		appendMessageTexts(&parts, message)
	}
	return parts
}

func extractOutputTexts(message *schema.Message) []string {
	if message == nil {
		return nil
	}

	parts := make([]string, 0, 2)
	appendMessageTexts(&parts, message)
	return parts
}

func appendMessageTexts(parts *[]string, message *schema.Message) {
	if message.Content != "" {
		*parts = append(*parts, message.Content)
	}
	for _, toolCall := range message.ToolCalls {
		if toolCall.Function.Name != "" {
			*parts = append(*parts, toolCall.Function.Name)
		}
		if toolCall.Function.Arguments != "" {
			*parts = append(*parts, toolCall.Function.Arguments)
		}
	}
	for _, part := range message.MultiContent {
		if part.Text != "" {
			*parts = append(*parts, part.Text)
		}
	}
	for _, part := range message.UserInputMultiContent {
		if part.Text != "" {
			*parts = append(*parts, part.Text)
		}
	}
	for _, part := range message.AssistantGenMultiContent {
		if part.Text != "" {
			*parts = append(*parts, part.Text)
		}
	}
}
