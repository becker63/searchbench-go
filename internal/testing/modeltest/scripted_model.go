package modeltest

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

var ErrNoScriptedResponses = errors.New("modeltest: no scripted responses remaining")

var _ model.ToolCallingChatModel = (*ScriptedModel)(nil)

// ScriptedResponse defines one deterministic model result.
type ScriptedResponse struct {
	Message *schema.Message
	Stream  []*schema.Message
	Err     error
}

// ScriptedCall records one Generate or Stream invocation.
type ScriptedCall struct {
	Method      string
	Messages    []*schema.Message
	OptionCount int
	Tools       []*schema.ToolInfo
}

type scriptedState struct {
	mu        sync.Mutex
	responses []ScriptedResponse
	calls     []ScriptedCall
}

// ScriptedModel is an in-process deterministic Eino-compatible chat model.
type ScriptedModel struct {
	state *scriptedState
	tools []*schema.ToolInfo
}

// NewScriptedModel returns a scripted model that serves responses in order.
func NewScriptedModel(responses ...ScriptedResponse) *ScriptedModel {
	cloned := make([]ScriptedResponse, len(responses))
	for i, response := range responses {
		cloned[i] = cloneScriptedResponse(response)
	}

	return &ScriptedModel{
		state: &scriptedState{
			responses: cloned,
		},
	}
}

// Generate returns the next scripted final response.
func (m *ScriptedModel) Generate(_ context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	response, err := m.recordAndDequeue("generate", input, opts)
	if err != nil {
		return nil, err
	}
	if response.Err != nil {
		return nil, response.Err
	}
	if response.Message == nil {
		return nil, nil
	}
	return cloneMessage(response.Message), nil
}

// Stream returns a deterministic message stream for the next scripted response.
func (m *ScriptedModel) Stream(_ context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	response, err := m.recordAndDequeue("stream", input, opts)
	if err != nil {
		return nil, err
	}
	if response.Err != nil {
		return nil, response.Err
	}

	chunks := response.Stream
	if len(chunks) == 0 && response.Message != nil {
		chunks = []*schema.Message{response.Message}
	}
	if len(chunks) == 0 {
		return schema.StreamReaderFromArray([]*schema.Message{}), nil
	}

	copied := make([]*schema.Message, len(chunks))
	for i, chunk := range chunks {
		copied[i] = cloneMessage(chunk)
	}
	return schema.StreamReaderFromArray(copied), nil
}

// WithTools returns a new view of this model with the supplied tools attached.
func (m *ScriptedModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	if m == nil {
		return nil, errors.New("modeltest: nil scripted model")
	}

	return &ScriptedModel{
		state: m.state,
		tools: cloneTools(tools),
	}, nil
}

// Calls returns a snapshot of all recorded model calls.
func (m *ScriptedModel) Calls() []ScriptedCall {
	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	calls := make([]ScriptedCall, len(m.state.calls))
	for i, call := range m.state.calls {
		calls[i] = ScriptedCall{
			Method:      call.Method,
			Messages:    cloneMessages(call.Messages),
			OptionCount: call.OptionCount,
			Tools:       cloneTools(call.Tools),
		}
	}
	return calls
}

func (m *ScriptedModel) recordAndDequeue(method string, input []*schema.Message, opts []model.Option) (ScriptedResponse, error) {
	if m == nil || m.state == nil {
		return ScriptedResponse{}, errors.New("modeltest: nil scripted model")
	}

	m.state.mu.Lock()
	defer m.state.mu.Unlock()

	m.state.calls = append(m.state.calls, ScriptedCall{
		Method:      method,
		Messages:    cloneMessages(input),
		OptionCount: len(opts),
		Tools:       cloneTools(m.tools),
	})

	if len(m.state.responses) == 0 {
		return ScriptedResponse{}, fmt.Errorf("%w for %s", ErrNoScriptedResponses, method)
	}

	response := m.state.responses[0]
	m.state.responses = m.state.responses[1:]
	return cloneScriptedResponse(response), nil
}

func cloneScriptedResponse(response ScriptedResponse) ScriptedResponse {
	cloned := ScriptedResponse{
		Err: response.Err,
	}
	if response.Message != nil {
		cloned.Message = cloneMessage(response.Message)
	}
	if len(response.Stream) > 0 {
		cloned.Stream = cloneMessages(response.Stream)
	}
	return cloned
}

func cloneMessages(messages []*schema.Message) []*schema.Message {
	if len(messages) == 0 {
		return nil
	}

	cloned := make([]*schema.Message, len(messages))
	for i, message := range messages {
		cloned[i] = cloneMessage(message)
	}
	return cloned
}

func cloneMessage(message *schema.Message) *schema.Message {
	if message == nil {
		return nil
	}

	cloned := *message

	if len(message.ToolCalls) > 0 {
		cloned.ToolCalls = append([]schema.ToolCall(nil), message.ToolCalls...)
	}
	if message.ResponseMeta != nil {
		responseMeta := *message.ResponseMeta
		cloned.ResponseMeta = &responseMeta
	}
	if len(message.MultiContent) > 0 {
		cloned.MultiContent = append([]schema.ChatMessagePart(nil), message.MultiContent...)
	}
	if len(message.UserInputMultiContent) > 0 {
		cloned.UserInputMultiContent = append([]schema.MessageInputPart(nil), message.UserInputMultiContent...)
	}
	if len(message.AssistantGenMultiContent) > 0 {
		cloned.AssistantGenMultiContent = append([]schema.MessageOutputPart(nil), message.AssistantGenMultiContent...)
	}
	if message.Extra != nil {
		cloned.Extra = make(map[string]any, len(message.Extra))
		for key, value := range message.Extra {
			cloned.Extra[key] = value
		}
	}

	return &cloned
}

func cloneTools(tools []*schema.ToolInfo) []*schema.ToolInfo {
	if len(tools) == 0 {
		return nil
	}

	cloned := make([]*schema.ToolInfo, len(tools))
	for i, tool := range tools {
		if tool == nil {
			continue
		}

		copyTool := *tool
		cloned[i] = &copyTool
	}
	return cloned
}
