// Package match is the vocabulary boundary for dataset items inside a round.
//
// Concrete types currently live in [github.com/becker63/searchbench-go/internal/pure/domain]
// to avoid import cycles with LCA localization helpers; this package re-exports
// those names so callers can depend on match.* without reaching into domain.
package match

import "github.com/becker63/searchbench-go/internal/pure/domain"

type (
	BenchmarkName = domain.BenchmarkName
	MatchID       = domain.MatchID
	MatchInput    = domain.MatchInput
	MatchOracle   = domain.MatchOracle
	MatchSpec     = domain.MatchSpec
)

const BenchmarkLCA = domain.BenchmarkLCA
