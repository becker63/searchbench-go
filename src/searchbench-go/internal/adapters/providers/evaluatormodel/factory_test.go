package evaluatormodel

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestNewFactoryFakeProviderUsesLocalFake(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	f := NewFactory(Config{Provider: "fake"})
	m, err := f(sampleSpec())
	if err != nil {
		t.Fatalf("factory(spec) error = %v", err)
	}
	msg, err := m.Generate(context.Background(), []*schema.Message{
		schema.UserMessage("hi"),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if msg == nil || len(msg.ToolCalls) == 0 {
		t.Fatalf("expected fake model tool call turn, got %#v", msg)
	}
}

func TestNewFactoryOpenAIWithoutCredentialsUsesLocalFake(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("CEREBRAS_API_KEY", "")

	f := NewFactory(Config{Provider: "openai", Model: "gpt-4o-mini"})
	m, err := f(sampleSpec())
	if err != nil {
		t.Fatalf("factory(spec) error = %v", err)
	}
	msg, err := m.Generate(context.Background(), []*schema.Message{
		schema.UserMessage("hi"),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if msg == nil || len(msg.ToolCalls) == 0 {
		t.Fatalf("expected fallback fake tool call turn, got %#v", msg)
	}
}

func TestNewFactoryOpenAIWithFixtureServer(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", "")

	server := modeltest.NewFakeOpenAIServer(modeltest.FakeResponse{
		Status: http.StatusOK,
		Body:   string(modeltest.MustFixture("chat_completion_success.json")),
	})
	t.Cleanup(server.Close)

	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_BASE_URL", server.BaseURL())

	f := NewFactory(Config{Provider: "openai", Model: "gpt-4o-mini"})
	m, err := f(sampleSpec())
	if err != nil {
		t.Fatalf("factory(spec) error = %v", err)
	}
	msg, err := m.Generate(context.Background(), []*schema.Message{
		schema.UserMessage("Say hello"),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if got, want := msg.Content, `{"files":["pkg/example.go"],"reasoning":"fixture"}`; got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
	reqs := server.Requests()
	if len(reqs) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(reqs))
	}
}

func TestNewFactoryOpenRouterUsesOpenRouterEnv(t *testing.T) {
	server := modeltest.NewFakeOpenAIServer(modeltest.FakeResponse{
		Status: http.StatusOK,
		Body:   string(modeltest.MustFixture("chat_completion_success.json")),
	})
	t.Cleanup(server.Close)

	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "router-key")
	t.Setenv("OPENROUTER_BASE_URL", server.BaseURL())

	f := NewFactory(Config{Provider: "openrouter", Model: "meta/llama"})
	m, err := f(sampleSpec())
	if err != nil {
		t.Fatalf("factory(spec) error = %v", err)
	}
	_, err = m.Generate(context.Background(), []*schema.Message{
		schema.UserMessage("x"),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(server.Requests()) != 1 {
		t.Fatalf("expected one HTTP request to fixture server")
	}
}

func sampleSpec() run.Spec {
	return run.Spec{
		ID: "incumbent-m1-sys",
		Match: domain.MatchSpec{
			ID: "m1",
		},
		System: domain.SystemSpec{
			Model: domain.ModelSpec{Provider: "openai", Name: "gpt-4o-mini"},
		},
	}
}
