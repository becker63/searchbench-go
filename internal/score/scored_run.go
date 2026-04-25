package score

import "github.com/becker63/searchbench-go/internal/run"

// ScoredRun is the minimum unit that can participate in a candidate report.
//
// It wraps a successful ExecutedRun with a complete ScoreSet.
type ScoredRun struct {
	Execution run.ExecutedRun `json:"execution"`
	Scores    ScoreSet        `json:"scores"`
}

// NewScoredRun constructs a scored run from an executed run and complete scores.
func NewScoredRun(execution run.ExecutedRun, scores ScoreSet) ScoredRun {
	return ScoredRun{
		Execution: execution,
		Scores:    scores,
	}
}
