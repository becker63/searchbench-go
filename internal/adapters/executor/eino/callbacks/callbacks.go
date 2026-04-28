// Package callbacks defines the Eino callback composition boundary for the
// SearchBench-Go evaluator execution layer.
//
// This package stays execution-local on purpose. It owns cold per-run callback
// construction and test fixtures for the evaluator seam. It does not own
// domain models, scoring, tracing systems, or a custom callback framework.
package callbacks

import (
	"context"
	"fmt"

	einocallbacks "github.com/cloudwego/eino/callbacks"
)

// Factory constructs one Eino callback handler for a single evaluator attempt.
//
// Construction must remain cold: factories may allocate or record test-local
// state, but they must not execute model calls or tool calls while building the
// handler set.
type Factory func(ctx context.Context) (einocallbacks.Handler, error)

// Config defines the callback factories to compose for one evaluator attempt.
type Config struct {
	Factories []Factory
}

// BuildCallbacks composes the configured callback handlers for one evaluator
// attempt.
//
// Factories are peers. BuildCallbacks preserves their order, skips nil
// factories, and fails closed if any factory returns an error.
func BuildCallbacks(ctx context.Context, config Config) ([]einocallbacks.Handler, error) {
	if len(config.Factories) == 0 {
		return nil, nil
	}

	handlers := make([]einocallbacks.Handler, 0, len(config.Factories))
	for i, factory := range config.Factories {
		if factory == nil {
			continue
		}

		handler, err := factory(ctx)
		if err != nil {
			return nil, fmt.Errorf("build callback %d: %w", i, err)
		}
		if handler != nil {
			handlers = append(handlers, handler)
		}
	}

	return handlers, nil
}
