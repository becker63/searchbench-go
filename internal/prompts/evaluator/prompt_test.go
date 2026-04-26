package evaluator

import (
	"context"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/domain"
)

func TestRenderPromptFromTypedInput(t *testing.T) {
	t.Parallel()

	task := domain.TaskSpec{
		ID:        domain.TaskID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("square/okhttp"),
			SHA:  domain.RepoSHA("abc123"),
		},
		Input: domain.TaskInput{
			Title: "Crash when retrying HTTP request",
			Body:  "The client crashes when the retry interceptor replays a request.",
		},
		Oracle: domain.TaskOracle{
			GoldFiles: []domain.RepoRelPath{"internal/should/not/leak.go"},
		},
	}

	prompt, err := Render(context.Background(), InputFromTask(task, []string{"fake_resolve"}))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"<searchbench-prompt>",
		"<role>",
		"<task>",
		"<repo>",
		"<issue>",
		"<available-tools>",
		"<output-contract>",
		"square/okhttp",
		"abc123",
		"Crash when retrying HTTP request",
		"retry interceptor",
		"fake_resolve",
		"predicted_files",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}

	for _, forbidden := range []string{
		"changed_files",
		"gold_files",
		"internal/should/not/leak.go",
	} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("prompt unexpectedly contains %q:\n%s", forbidden, prompt)
		}
	}
}
