package compare

import (
	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

type Results struct {
	runs        domain.Pair[[]score.ScoredRun]
	failures    domain.Pair[[]run.RunFailure]
	regressions []report.Regression
}

type Summary struct {
	Runs        domain.Pair[[]score.ScoredRun]
	Failures    domain.Pair[[]run.RunFailure]
	Comparisons []report.ScoreComparison
	Regressions []report.Regression
}

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

func (r *Results) AddTaskResult(result TaskComparisonResult) {
	if result.Runs.Baseline != nil {
		r.runs.Baseline = append(r.runs.Baseline, *result.Runs.Baseline)
	}
	if result.Runs.Candidate != nil {
		r.runs.Candidate = append(r.runs.Candidate, *result.Runs.Candidate)
	}
	if result.Failures.Baseline != nil {
		r.failures.Baseline = append(r.failures.Baseline, *result.Failures.Baseline)
	}
	if result.Failures.Candidate != nil {
		r.failures.Candidate = append(r.failures.Candidate, *result.Failures.Candidate)
	}
	r.regressions = append(r.regressions, result.Regressions...)
}

func (r Results) Summary() Summary {
	metricComparisons := score.CompareAverages(r.runs.Baseline, r.runs.Candidate)
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
