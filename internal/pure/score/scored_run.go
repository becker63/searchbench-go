package score

import run "github.com/becker63/searchbench-go/internal/pure/execution"

// ScoredRun is the minimum unit that can participate in a round report.
//
// It wraps a successful ExecutedRun with a complete validated ScoreSet. If a
// required metric cannot be computed, scoring should fail instead of
// constructing ScoredRun.
type ScoredRun struct {
	Execution run.ExecutedRun `json:"execution"`
	Scores    ScoreSet        `json:"scores"`
}

// NewScoredRun constructs a scored run from an executed run and a complete
// required score set.
func NewScoredRun(execution run.ExecutedRun, scores ScoreSet) (ScoredRun, error) {
	if err := scores.Validate(); err != nil {
		return ScoredRun{}, err
	}
	return ScoredRun{
		Execution: execution,
		Scores:    scores,
	}, nil
}
