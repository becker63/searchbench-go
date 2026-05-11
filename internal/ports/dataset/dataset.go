// Package dataset declares the MatchSource port used by rounds to obtain the
// list of MatchSpec values for a run without binding to a concrete dataset
// adapter (Hugging Face datasets, local files, deterministic fakes, ...).
package dataset

import (
	"context"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// Request is the dataset-agnostic description of which matches a round needs.
//
// It mirrors the manifest's `dataset` block so adapters can route on Kind/Name
// while still receiving the manifest directory for local resolution.
type Request struct {
	ManifestDir string
	Kind        string
	Name        string
	Config      string
	Split       string
	MaxItems    *int
}

// MatchSource is the port rounds use to materialize the match list. Adapters
// must return at least one match.
type MatchSource interface {
	Matches(ctx context.Context, req Request) (domain.NonEmpty[domain.MatchSpec], error)
}
