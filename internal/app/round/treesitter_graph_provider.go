package round

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	tsgraph "github.com/becker63/searchbench-go/internal/adapters/codegraph/treesitter"
	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// TreesitterGraphProvider builds code graphs from materialized repositories when
// [domain.MatchSpec.Repo.Path] is non-empty and tree-sitter indexing succeeds.
// On missing paths, CGO-disabled stubs, or indexer failures it delegates to
// Fallback so offline CI and fake-local rounds stay deterministic.
type TreesitterGraphProvider struct {
	Fallback compare.GraphProvider
}

// NewTreesitterGraphProvider returns a non-nil provider using fallback when the
// repository path is unusable.
func NewTreesitterGraphProvider(fallback compare.GraphProvider) *TreesitterGraphProvider {
	return &TreesitterGraphProvider{Fallback: fallback}
}

// GraphForTask satisfies [compare.GraphProvider].
func (p *TreesitterGraphProvider) GraphForTask(ctx context.Context, task domain.MatchSpec) (codegraph.Graph, error) {
	if p == nil || p.Fallback == nil {
		return nil, errors.New("round: TreesitterGraphProvider requires non-nil Fallback")
	}
	root := strings.TrimSpace(string(task.Repo.Path))
	if root == "" {
		return p.Fallback.GraphForTask(ctx, task)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return p.Fallback.GraphForTask(ctx, task)
	}
	store, err := tsgraph.BuildDirectory(absRoot)
	if err != nil {
		if errors.Is(err, tsgraph.ErrCGODisabled) {
			return p.Fallback.GraphForTask(ctx, task)
		}
		return p.Fallback.GraphForTask(ctx, task)
	}
	g, err := store.Build()
	if err != nil {
		return p.Fallback.GraphForTask(ctx, task)
	}
	return g, nil
}
