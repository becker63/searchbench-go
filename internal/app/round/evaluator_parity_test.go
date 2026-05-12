package round

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/agents/evaluator/prompt"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

func TestFilterToolsByAllowSetExcludesDeniedNames(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	spec := run.NewSpec(
		domain.RunID("challenger-m1-sys"),
		domain.MatchSpec{
			ID: domain.MatchID("m1"),
			Input: domain.MatchInput{
				Title: "t",
			},
		},
		domain.SystemSpec{},
	)
	tools, err := fake.ToolFactory(spec)
	if err != nil {
		t.Fatal(err)
	}
	allow := map[string]struct{}{
		"expand": {},
	}
	filtered, err := filterToolsByAllowSet(ctx, tools, allow)
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	info, err := filtered[0].Info(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "expand" {
		t.Fatalf("tool name = %q, want expand", info.Name)
	}
}

func TestFilterToolNamesMatchEffectiveAllowed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resolved, _ := newTestRound(t)
	effective := resolved.Round.Evaluator.ToolPolicy.EffectiveAllowed
	tools, err := fake.ToolFactory(sampleSpec())
	if err != nil {
		t.Fatal(err)
	}
	allow := make(map[string]struct{}, len(effective))
	for _, name := range effective {
		allow[name] = struct{}{}
	}
	filtered, err := filterToolsByAllowSet(ctx, tools, allow)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tt := range filtered {
		info, err := tt.Info(ctx)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, info.Name)
	}
	slices.Sort(got)
	want := append([]string(nil), effective...)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("registered tool names = %v, want %v", got, want)
	}
}

func TestRenderedPromptListsOnlyEffectiveTools(t *testing.T) {
	t.Parallel()

	task := domain.MatchSpec{
		ID:        domain.MatchID("id"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("r"),
			SHA:  domain.RepoSHA("s"),
		},
		Input: domain.MatchInput{Title: "t", Body: "b"},
	}
	effective := []string{"expand", "resolve"}
	promptStr, err := prompt.Render(
		context.Background(),
		prompt.InputFromMatch(task, effective, "extra guidance"),
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"resolve_and_expand"} {
		if strings.Contains(promptStr, "<tool>"+forbidden+"</tool>") {
			t.Fatalf("prompt should not list %q", forbidden)
		}
	}
	if !strings.Contains(promptStr, "<tool>expand</tool>") || !strings.Contains(promptStr, "<tool>resolve</tool>") {
		t.Fatalf("prompt missing effective tools:\n%s", promptStr)
	}
	if !strings.Contains(promptStr, "<system-prompt>") || !strings.Contains(promptStr, "extra guidance") {
		t.Fatal("system prompt missing")
	}
	if !strings.Contains(promptStr, "You are the SearchBench evaluator agent") {
		t.Fatal("base role missing")
	}
	if !strings.Contains(promptStr, "strict JSON") {
		t.Fatal("output contract missing")
	}
}

func TestResolvedRoundIncludesEvaluatorToolPolicyFields(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
	resolved, _ := newTestRound(t)
	tp := resolved.Round.Evaluator.ToolPolicy
	if len(tp.EffectiveAllowed) == 0 {
		t.Fatal("empty effective allowed")
	}
	if tp.PolicySHA256 == "" {
		t.Fatal("empty policy hash")
	}
}

func sampleSpec() run.Spec {
	task := domain.MatchSpec{
		ID: domain.MatchID("m1"),
		Input: domain.MatchInput{
			Title: "title",
			Body:  "body",
		},
	}
	return run.NewSpec(domain.RunID("run-1"), task, domain.SystemSpec{})
}
