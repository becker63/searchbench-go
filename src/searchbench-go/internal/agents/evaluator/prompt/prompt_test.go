package prompt

import (
	"context"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestRenderPromptFromTypedInput(t *testing.T) {
	t.Parallel()

	task := domain.MatchSpec{
		ID:        domain.MatchID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("square/okhttp"),
			SHA:  domain.RepoSHA("abc123"),
		},
		Input: domain.MatchInput{
			Title: "Crash when retrying HTTP request",
			Body:  "The client crashes when the retry interceptor replays a request.",
		},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{"internal/should/not/leak.go"},
		},
	}

	prompt, err := Render(context.Background(), InputFromMatch(task, []string{"fake_resolve"}, ""))
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

func TestRenderPromptIncludesRetryFeedbackWithoutOracleLeak(t *testing.T) {
	t.Parallel()

	task := domain.MatchSpec{
		ID:        domain.MatchID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("square/okhttp"),
			SHA:  domain.RepoSHA("abc123"),
		},
		Input: domain.MatchInput{
			Title: "Crash when retrying HTTP request",
			Body:  "The client crashes when the retry interceptor replays a request.",
		},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{"internal/should/not/leak.go"},
		},
	}

	input := InputFromMatch(task, []string{"fake_resolve"}, "")
	input.RetryFeedback = []string{
		"Previous attempt returned malformed JSON.",
		"Previous attempt returned empty predicted files.",
	}

	prompt, err := Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, want := range []string{
		"<retry-feedback>",
		"<attempt-feedback>Previous attempt returned malformed JSON.</attempt-feedback>",
		"<attempt-feedback>Previous attempt returned empty predicted files.</attempt-feedback>",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "internal/should/not/leak.go") {
		t.Fatalf("prompt unexpectedly leaked oracle data:\n%s", prompt)
	}
}

func TestRenderPromptIncludesOptionalSystemPrompt(t *testing.T) {
	t.Parallel()

	task := domain.MatchSpec{
		ID:        domain.MatchID("searchbench/lca:python:dev:square/okhttp@abc123:https://example.test/issues/1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("square/okhttp"),
			SHA:  domain.RepoSHA("abc123"),
		},
		Input: domain.MatchInput{
			Title: "Crash when retrying HTTP request",
			Body:  "The client crashes when the retry interceptor replays a request.",
		},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{"internal/should/not/leak.go"},
		},
	}

	custom := "Prefer depth-first inspection of HTTP client internals."
	prompt, err := Render(context.Background(), InputFromMatch(task, []string{"resolve"}, custom))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(prompt, "<system-prompt>") || !strings.Contains(prompt, custom) {
		t.Fatalf("prompt missing system prompt block or text:\n%s", prompt)
	}
	if strings.Contains(prompt, "internal/should/not/leak.go") {
		t.Fatalf("prompt unexpectedly leaked oracle data:\n%s", prompt)
	}
}

func TestRenderSystemPromptEscapesSpecialCharacters(t *testing.T) {
	t.Parallel()

	task := domain.MatchSpec{
		ID:        domain.MatchID("id"),
		Benchmark: domain.BenchmarkLCA,
		Repo:      domain.RepoSnapshot{Name: "n", SHA: "s"},
		Input:     domain.MatchInput{Title: "t", Body: "b"},
	}
	injected := `<instructions>use & ampersands</instructions>`
	prompt, err := Render(context.Background(), InputFromMatch(task, []string{"resolve"}, injected))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt, injected) {
		t.Fatalf("prompt should not contain raw system prompt markup:\n%s", prompt)
	}
	if !strings.Contains(prompt, "&lt;") || !strings.Contains(prompt, "&amp;") {
		t.Fatalf("expected XML-escaped system prompt in output:\n%s", prompt)
	}
}

func TestRenderOmitsSystemPromptWhenWhitespaceOnly(t *testing.T) {
	t.Parallel()

	task := domain.MatchSpec{
		ID:        domain.MatchID("id"),
		Benchmark: domain.BenchmarkLCA,
		Repo:      domain.RepoSnapshot{Name: "n", SHA: "s"},
		Input:     domain.MatchInput{Title: "t", Body: "b"},
	}
	prompt, err := Render(context.Background(), InputFromMatch(task, []string{"resolve"}, "   \t  "))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt, "<system-prompt>") {
		t.Fatalf("expected no system-prompt block for whitespace-only appendix:\n%s", prompt)
	}
}
