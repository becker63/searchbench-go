// Package fake is a deterministic GraphProvider used by local-only fake rounds.
//
// It satisfies the structural compare.GraphProvider interface without binding
// to that internal package, so rounds compose it as the default graph provider
// when no real codegraph backend is configured.
package fake

import (
	"context"

	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// GraphProvider returns a single-file, single-function graph keyed off the
// match's first gold file. It is the deterministic default used by rounds
// running in local-fake mode.
type GraphProvider struct{}

// New constructs a GraphProvider. It exists to match the convention of the
// other fake adapters and keep call sites consistent.
func New() GraphProvider { return GraphProvider{} }

// GraphForTask satisfies the structural compare.GraphProvider interface.
func (GraphProvider) GraphForTask(_ context.Context, task domain.MatchSpec) (codegraph.Graph, error) {
	store := codegraph.NewStore()
	fileID := codegraph.NodeID("file-" + task.ID.String())
	fnID := codegraph.NodeID("fn-" + task.ID.String())

	if err := store.AddNode(codegraph.NewFileNode(fileID, task.Oracle.GoldFiles[0])); err != nil {
		return nil, err
	}
	if err := store.AddNode(codegraph.NewFunctionNode(fnID, task.Oracle.GoldFiles[0], "score", 1, 10)); err != nil {
		return nil, err
	}
	if err := store.AddEdge(codegraph.Edge{
		From:   fileID,
		To:     fnID,
		Kind:   codegraph.EdgeContains,
		Weight: 1,
	}); err != nil {
		return nil, err
	}
	return store.Build()
}
