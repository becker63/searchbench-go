package evaluatormodel

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// stripReasoningModel wraps a chat model and removes reasoning_content fields from
// outbound requests. Cerebras chat completions reject OpenAI-style reasoning payloads.
type stripReasoningModel struct {
	inner model.ToolCallingChatModel
}

func wrapCerebrasCompat(m model.ToolCallingChatModel) model.ToolCallingChatModel {
	if m == nil {
		return nil
	}
	return &stripReasoningModel{inner: m}
}

func (m *stripReasoningModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	return m.inner.Generate(ctx, stripReasoningMessages(input), opts...)
}

func (m *stripReasoningModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return m.inner.Stream(ctx, stripReasoningMessages(input), opts...)
}

func (m *stripReasoningModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	next, err := m.inner.WithTools(tools)
	if err != nil {
		return nil, err
	}
	return wrapCerebrasCompat(next), nil
}

func stripReasoningMessages(input []*schema.Message) []*schema.Message {
	if len(input) == 0 {
		return input
	}
	out := make([]*schema.Message, len(input))
	for i, msg := range input {
		if msg == nil {
			continue
		}
		cp := *msg
		cp.ReasoningContent = ""
		if cp.Extra != nil {
			delete(cp.Extra, "reasoning-content")
		}
		out[i] = &cp
	}
	return out
}
