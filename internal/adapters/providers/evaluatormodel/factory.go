package evaluatormodel

import (
	"context"
	"os"
	"strings"

	openai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"

	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

// Config is manifest evaluator model configuration (provider-neutral).
type Config struct {
	Provider        string
	Model           string
	MaxOutputTokens int
}

// NewFactory returns a factory that honors [Config] and the environment.
//
// When credentials for a cloud provider are absent, the returned factory is the
// same as [evaluatorfake.ModelFactory] so default validation runs need no secrets.
func NewFactory(cfg Config) func(spec run.Spec) (model.ToolCallingChatModel, error) {
	switch normalizeProvider(cfg.Provider) {
	case "", "fake":
		return evaluatorfake.ModelFactory
	case "openai", "openrouter", "cerebras":
		m, ok := newOpenAIChatModel(cfg)
		if !ok {
			return evaluatorfake.ModelFactory
		}
		return func(run.Spec) (model.ToolCallingChatModel, error) {
			return m, nil
		}
	default:
		return evaluatorfake.ModelFactory
	}
}

func normalizeProvider(p string) string {
	return strings.ToLower(strings.TrimSpace(p))
}

type openAICreds struct {
	apiKey  string
	baseURL string
}

func resolveOpenAICreds(provider string) openAICreds {
	switch provider {
	case "openai":
		return openAICreds{
			apiKey:  strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
			baseURL: strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")),
		}
	case "openrouter":
		key := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
		if key == "" {
			key = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
		}
		base := strings.TrimSpace(os.Getenv("OPENROUTER_BASE_URL"))
		if base == "" {
			base = "https://openrouter.ai/api/v1"
		}
		return openAICreds{apiKey: key, baseURL: base}
	case "cerebras":
		base := strings.TrimSpace(os.Getenv("CEREBRAS_BASE_URL"))
		if base == "" {
			base = "https://api.cerebras.ai/v1"
		}
		return openAICreds{
			apiKey:  strings.TrimSpace(os.Getenv("CEREBRAS_API_KEY")),
			baseURL: base,
		}
	default:
		return openAICreds{}
	}
}

func newOpenAIChatModel(cfg Config) (model.ToolCallingChatModel, bool) {
	p := normalizeProvider(cfg.Provider)
	creds := resolveOpenAICreds(p)
	if creds.apiKey == "" {
		return nil, false
	}

	modelName := strings.TrimSpace(cfg.Model)
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	ocfg := &openai.ChatModelConfig{
		APIKey: creds.apiKey,
		Model:  modelName,
	}
	if creds.baseURL != "" {
		ocfg.BaseURL = creds.baseURL
	}
	if cfg.MaxOutputTokens > 0 {
		v := cfg.MaxOutputTokens
		ocfg.MaxCompletionTokens = &v
	}

	m, err := openai.NewChatModel(context.Background(), ocfg)
	if err != nil {
		return nil, false
	}
	return m, true
}
