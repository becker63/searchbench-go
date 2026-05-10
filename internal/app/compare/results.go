package compare

import (
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

type Results struct {
	runs        domain.Pair[[]score.ScoredRun]
	failures    domain.Pair[[]run.RunFailure]
	regressions []report.Regression
}

// Summary is the aggregated comparison view produced after all task work has
// finished.
type Summary struct {
	Runs        domain.Pair[[]score.ScoredRun]
	Failures    domain.Pair[[]run.RunFailure]
	Comparisons []report.ScoreComparison
	Regressions []report.Regression
}

// NewResults constructs the single-threaded comparison accumulator used by
// Runner after task work has been collected in deterministic order.
func NewResults(capacity int) Results {
	return Results{
		runs: domain.NewPair(
			make([]score.ScoredRun, 0, capacity),
			make([]score.ScoredRun, 0, capacity),
		),
		failures: domain.NewPair(
			make([]run.RunFailure, 0),
			make([]run.RunFailure, 0),
		),
		regressions: make([]report.Regression, 0),
	}
}

// AddTaskResult appends one task comparison outcome into the aggregate result.
//
// Results is intentionally single-threaded. Runner collects task work first and
// then feeds the accumulator on the main goroutine to preserve deterministic
// ordering without locks.
func (r *Results) AddTaskResult(result TaskComparisonResult) {
	if result.Runs.Incumbent != nil {
		r.runs.Incumbent = append(r.runs.Incumbent, *result.Runs.Incumbent)
	}
	if result.Runs.Challenger != nil {
		r.runs.Challenger = append(r.runs.Challenger, *result.Runs.Challenger)
	}
	if result.Failures.Incumbent != nil {
		r.failures.Incumbent = append(r.failures.Incumbent, *result.Failures.Incumbent)
	}
	if result.Failures.Challenger != nil {
		r.failures.Challenger = append(r.failures.Challenger, *result.Failures.Challenger)
	}
	r.regressions = append(r.regressions, result.Regressions...)
}

// Summary reduces the accumulated runs and failures into report-facing
// comparisons plus the collected regressions.
func (r Results) Summary() Summary {
	metricComparisons := score.CompareAverages(r.runs.Incumbent, r.runs.Challenger)
	comparisons := make([]report.ScoreComparison, 0, len(metricComparisons))
	for _, comparison := range metricComparisons {
		comparisons = append(comparisons, report.NewScoreComparisonFromMetric(comparison))
	}

	return Summary{
		Runs:        r.runs,
		Failures:    r.failures,
		Comparisons: comparisons,
		Regressions: append([]report.Regression(nil), r.regressions...),
	}
}
