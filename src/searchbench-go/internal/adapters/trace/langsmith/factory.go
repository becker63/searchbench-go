// Package langsmith is the SearchBench adapter boundary for optional LangSmith/Eino
// trace callbacks. SearchBench bundle, scoring, evidence, and decisions remain
// authoritative; this integration is observational only and must not define core
// domain types.
package langsmith

import (
	"context"
	"os"
	"strings"

	extlangsmith "github.com/cloudwego/eino-ext/callbacks/langsmith"
	einocallbacks "github.com/cloudwego/eino/callbacks"
)

// Well-known environment variables for optional LangSmith export (same names as
// LangChain/LangSmith tooling when applicable).
const (
	EnvAPIKey = "LANGSMITH_API_KEY"
	EnvAPIURL = "LANGSMITH_API_URL"
	// EnvSession overrides the default LangSmith session name for trace grouping.
	EnvSession = "SEARCHBENCH_LANGSMITH_SESSION"
)

// Config configures construction of the upstream LangSmith Eino callback handler.
type Config struct {
	APIKey string
	APIURL string
	// RunIDGen is forwarded to the upstream handler; nil uses its default.
	RunIDGen func(context.Context) string
}

// HandlerFactory matches the evaluator callback composer shape
// (internal/agents/evaluator/eino/callbacks.Factory).
type HandlerFactory func(context.Context) (einocallbacks.Handler, error)

// NewHandlerFactory returns a cold factory that yields the LangChain LangSmith
// Eino handler, or (nil, nil) when APIKey is empty so tracing stays off without
// credentials.
func NewHandlerFactory(cfg Config) (HandlerFactory, error) {
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" {
		return nil, nil
	}
	upstream := &extlangsmith.Config{
		APIKey:   key,
		APIURL:   strings.TrimSpace(cfg.APIURL),
		RunIDGen: cfg.RunIDGen,
	}
	h, err := extlangsmith.NewLangsmithHandler(upstream)
	if err != nil {
		return nil, err
	}
	return func(context.Context) (einocallbacks.Handler, error) {
		return h, nil
	}, nil
}

// HandlerFactoryFromEnv builds a factory from LANGSMITH_API_KEY /
// LANGSMITH_API_URL. When the API key is unset, returns (nil, nil).
func HandlerFactoryFromEnv() (HandlerFactory, error) {
	return NewHandlerFactory(Config{
		APIKey: os.Getenv(EnvAPIKey),
		APIURL: os.Getenv(EnvAPIURL),
	})
}
