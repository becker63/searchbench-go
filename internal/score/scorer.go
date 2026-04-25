package score

import "context"

// Engine computes the complete required ScoreSet for one executed run.
//
// If a required metric cannot be computed, Score should return an error rather
// than a partial score set.
type Engine interface {
	Score(ctx context.Context, input Input) (ScoreSet, error)
}
