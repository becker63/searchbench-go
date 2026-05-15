package langsmith

import (
	"context"
	"testing"
)

func TestNewHandlerFactory_DisabledWithoutAPIKey(t *testing.T) {
	t.Parallel()

	f, err := NewHandlerFactory(Config{})
	if err != nil {
		t.Fatalf("NewHandlerFactory() error = %v", err)
	}
	if f != nil {
		t.Fatal("expected nil factory without API key")
	}
}

func TestHandlerFactoryFromEnv_DefaultOffline(t *testing.T) {
	t.Setenv(EnvAPIKey, "")
	f, err := HandlerFactoryFromEnv()
	if err != nil {
		t.Fatalf("HandlerFactoryFromEnv() error = %v", err)
	}
	if f != nil {
		t.Fatal("expected nil factory with empty env")
	}
}

func TestNewHandlerFactory_EnabledWithAPIKey(t *testing.T) {
	t.Parallel()

	f, err := NewHandlerFactory(Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("NewHandlerFactory() error = %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil factory")
	}
	h, err := f(context.Background())
	if err != nil {
		t.Fatalf("factory() error = %v", err)
	}
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestAugmentContext_NoKeyNoOp(t *testing.T) {
	t.Setenv(EnvAPIKey, "")
	ctx := context.Background()
	out := AugmentContext(ctx, ContextLabels{MatchID: "m1"})
	if out != ctx {
		t.Fatal("expected same context when tracing disabled")
	}
}

func TestAugmentContext_WithKeyAddsTrace(t *testing.T) {
	t.Setenv(EnvAPIKey, "k")
	ctx := context.Background()
	out := AugmentContext(ctx, ContextLabels{
		SessionName: "sess",
		MatchID:     "mid",
		RunID:       "rid",
		Role:        "incumbent",
	})
	if out == ctx {
		t.Fatal("expected derived context when tracing enabled")
	}
}
