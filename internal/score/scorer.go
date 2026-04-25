package score

import "context"

// Engine computes the complete required ScoreSet for one executed run.
type Engine interface {
	Score(ctx context.Context, input Input) (ScoreSet, error)
}
