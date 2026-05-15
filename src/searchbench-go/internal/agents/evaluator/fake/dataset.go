package fake

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/becker63/searchbench-go/internal/ports/dataset"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// MatchSource implements dataset.MatchSource with one deterministic match.
type MatchSource struct{}

// NewMatchSource constructs a MatchSource.
func NewMatchSource() MatchSource { return MatchSource{} }

// Matches satisfies dataset.MatchSource.
func (MatchSource) Matches(_ context.Context, req dataset.Request) (domain.NonEmpty[domain.MatchSpec], error) {
	repoPath := domain.HostPath(filepath.Join(req.ManifestDir, "fake-repo"))
	task := domain.MatchSpec{
		ID:        domain.MatchID("local-fake-match-1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("searchbench/local-fake"),
			SHA:  domain.RepoSHA("0000000"),
			Path: repoPath,
		},
		Input: domain.MatchInput{
			Title: fmt.Sprintf("Fake %s/%s match", req.Config, req.Split),
			Body:  "This deterministic local match exists only to prove manifest-driven composition.",
		},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{"src/search_target.go"},
		},
	}
	return domain.NewNonEmpty(task), nil
}
