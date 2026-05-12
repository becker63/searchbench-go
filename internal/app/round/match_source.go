package round

import (
	"context"

	lcaadapter "github.com/becker63/searchbench-go/internal/adapters/dataset/lca"
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/ports/dataset"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

type compositeMatchSource struct {
	jetBrainsLCA dataset.MatchSource
	localFake    dataset.MatchSource
}

func newDefaultMatchSource() dataset.MatchSource {
	return compositeMatchSource{
		jetBrainsLCA: lcaadapter.NewSource(),
		localFake:    evaluatorfake.NewMatchSource(),
	}
}

func (c compositeMatchSource) Matches(ctx context.Context, req dataset.Request) (domain.NonEmpty[domain.MatchSpec], error) {
	if lcaadapter.IsJetBrainsDataset(req) {
		return c.jetBrainsLCA.Matches(ctx, req)
	}
	return c.localFake.Matches(ctx, req)
}
