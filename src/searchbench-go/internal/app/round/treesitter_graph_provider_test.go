package round

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	tsgraph "github.com/becker63/searchbench-go/internal/adapters/codegraph/treesitter"
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestTreesitterGraphProvider_DelegatesWhenRepoPathEmpty(t *testing.T) {
	t.Parallel()

	fallback := evaluatorfake.NewGraphProvider()
	p := NewTreesitterGraphProvider(fallback)
	ctx := context.Background()
	task := domain.MatchSpec{
		ID:        domain.MatchID("t1"),
		Benchmark: domain.BenchmarkLCA,
		Repo:      domain.RepoSnapshot{Name: "n", SHA: "sha"},
		Oracle:    domain.MatchOracle{GoldFiles: []domain.RepoRelPath{"x.go"}},
	}

	gWant, err := fallback.GraphForTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	gGot, err := p.GraphForTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	if len(gGot.Nodes()) != len(gWant.Nodes()) {
		t.Fatalf("nodes got=%d want=%d", len(gGot.Nodes()), len(gWant.Nodes()))
	}
}

func TestTreesitterGraphProvider_RealIndexWhenAvailable(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	mainGo := filepath.Join(tmp, "main.go")
	if err := os.WriteFile(mainGo, []byte(`package main
func A() {}
func B() { A() }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := tsgraph.BuildDirectory(tmp); err != nil {
		if errors.Is(err, tsgraph.ErrCGODisabled) {
			t.Skip("CGO disabled; tree-sitter indexer unavailable")
		}
		t.Fatalf("BuildDirectory: %v", err)
	}

	fallback := evaluatorfake.NewGraphProvider()
	p := NewTreesitterGraphProvider(fallback)
	ctx := context.Background()
	task := domain.MatchSpec{
		ID:        domain.MatchID("t2"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: "n",
			SHA:  "sha",
			Path: domain.HostPath(tmp),
		},
		Oracle: domain.MatchOracle{GoldFiles: []domain.RepoRelPath{"main.go"}},
	}

	gFake, err := fallback.GraphForTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	gGot, err := p.GraphForTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}
	if len(gGot.Nodes()) <= len(gFake.Nodes()) {
		t.Fatalf("expected richer treesitter graph; got %d nodes, fake has %d", len(gGot.Nodes()), len(gFake.Nodes()))
	}
}
