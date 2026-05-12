package round

import (
	"context"
	"testing"
	"time"

	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestLocalizationGraphScorer_OverridesHopMetricsWhenPathExists(t *testing.T) {
	t.Parallel()

	fileA := domain.RepoRelPath("a.go")
	fileB := domain.RepoRelPath("b.go")

	store := codegraph.NewStore()
	fA := codegraph.NewFunctionNode("fa", fileA, "fa", 1, 5)
	fB := codegraph.NewFunctionNode("fb", fileB, "fb", 1, 5)
	if err := store.AddNode(fA); err != nil {
		t.Fatal(err)
	}
	if err := store.AddNode(fB); err != nil {
		t.Fatal(err)
	}
	if err := store.AddEdge(codegraph.NewEdge("fb", "fa", codegraph.EdgeCalls)); err != nil {
		t.Fatal(err)
	}
	g, err := store.Build()
	if err != nil {
		t.Fatal(err)
	}

	match := domain.MatchSpec{
		ID:        domain.MatchID("m1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: "demo",
			SHA:  "deadbeef",
			Path: "",
		},
		Oracle: domain.MatchOracle{GoldFiles: []domain.RepoRelPath{fileA}},
	}
	sys := domain.SystemSpec{
		ID:           domain.SystemID("inc"),
		Name:         "inc",
		Backend:      domain.BackendIterativeContext,
		Model:        domain.ModelSpec{Provider: "openai", Name: "gpt-4o-mini"},
		PromptBundle: domain.PromptBundleRef{Name: "default"},
	}
	if err := sys.Validate(); err != nil {
		t.Fatal(err)
	}

	spec := run.NewSpec(domain.RunID("run-1"), match, sys)
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("sess"))
	ts := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
	executed := run.NewExecuted(
		prepared,
		domain.Prediction{Files: []domain.RepoRelPath{fileB}},
		domain.UsageSummary{},
		domain.TraceID(""),
		ts,
		ts.Add(time.Minute),
	)

	scorer := NewLocalizationGraphScorer(evaluatorfake.NewScorer())
	got, err := scorer.Score(context.Background(), score.Input{
		Run:   executed,
		Graph: g,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.GoldHop.Value != 1 || got.IssueHop.Value != 1 {
		t.Fatalf("hop metrics = gold=%v issue=%v; want 1/1", got.GoldHop.Value, got.IssueHop.Value)
	}
	if got.TokenEfficiency.Value != 0.85 {
		t.Fatalf("expected fallback token efficiency preserved; got %v", got.TokenEfficiency.Value)
	}
}
