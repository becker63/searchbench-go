package callbacks

import (
	"context"
	"strings"

	einocallbacks "github.com/cloudwego/eino/callbacks"
	einocomponents "github.com/cloudwego/eino/components"
	einomodel "github.com/cloudwego/eino/components/model"

	"github.com/becker63/searchbench-go/internal/pure/usage"
)

// HashRegistryConfig configures request/response hashing for live attribution (#86).
type HashRegistryConfig struct {
	Registry *usage.HashRegistry
}

// HashRegistryCallback records deterministic model payload hashes.
type HashRegistryCallback struct {
	registry *usage.HashRegistry
	handler  einocallbacks.Handler
}

// NewHashRegistryCallback constructs a cold hash-registry callback.
func NewHashRegistryCallback(config HashRegistryConfig) (*HashRegistryCallback, error) {
	callback := &HashRegistryCallback{registry: config.Registry}
	callback.handler = einocallbacks.NewHandlerBuilder().
		OnStartFn(callback.onStart).
		OnEndFn(callback.onEnd).
		Build()
	return callback, nil
}

// NewHashRegistryCallbackFactory returns a Factory that attaches to reg each attempt.
func NewHashRegistryCallbackFactory(reg *usage.HashRegistry) Factory {
	return func(context.Context) (einocallbacks.Handler, error) {
		cb, err := NewHashRegistryCallback(HashRegistryConfig{Registry: reg})
		if err != nil {
			return nil, err
		}
		return cb.Handler(), nil
	}
}

// Handler returns the Eino callback handler.
func (c *HashRegistryCallback) Handler() einocallbacks.Handler {
	if c == nil {
		return nil
	}
	return c.handler
}

func (c *HashRegistryCallback) onStart(ctx context.Context, info *einocallbacks.RunInfo, input einocallbacks.CallbackInput) context.Context {
	if c == nil || c.registry == nil || info == nil || info.Component != einocomponents.ComponentOfChatModel {
		return ctx
	}
	modelInput := einomodel.ConvCallbackInput(input)
	if modelInput == nil {
		return ctx
	}
	c.registry.RecordRequest([]byte(strings.Join(extractInputTexts(modelInput.Messages), "\n")))
	return ctx
}

func (c *HashRegistryCallback) onEnd(ctx context.Context, info *einocallbacks.RunInfo, output einocallbacks.CallbackOutput) context.Context {
	if c == nil || c.registry == nil || info == nil || info.Component != einocomponents.ComponentOfChatModel {
		return ctx
	}
	modelOutput := einomodel.ConvCallbackOutput(output)
	if modelOutput == nil || modelOutput.Message == nil {
		return ctx
	}
	c.registry.RecordResponse([]byte(strings.Join(extractOutputTexts(modelOutput.Message), "\n")))
	return ctx
}
