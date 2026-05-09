package report

import (
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// RegressionSeverity describes how serious a candidate regression is.
type RegressionSeverity string

const (
	RegressionMinor    RegressionSeverity = "minor"
	RegressionBlocking RegressionSeverity = "blocking"
)

// Regression records a metric/match-level reason the challenger regressed.
type Regression struct {
	MatchID domain.MatchID   `json:"match_id"`
	Metric  score.MetricName `json:"metric"`

	Baseline  float64 `json:"baseline"`
	Candidate float64 `json:"candidate"`
	Delta     float64 `json:"delta"`

	Severity RegressionSeverity `json:"severity"`
	Reason   string             `json:"reason"`
}

// NewRegressionFromMetricComparison converts a score-level regression into a
// report-facing regression record.
func NewRegressionFromMetricComparison(
	matchID domain.MatchID,
	comparison score.MetricComparison,
	severity RegressionSeverity,
	reason string,
) Regression {
	return Regression{
		MatchID:   matchID,
		Metric:    comparison.Metric,
		Baseline:  comparison.Baseline,
		Candidate: comparison.Candidate,
		Delta:     comparison.Delta,
		Severity:  severity,
		Reason:    reason,
	}
}

// RegressionsForMatch converts metric comparisons into report-level regressions
// for one match.
func RegressionsForMatch(matchID domain.MatchID, comparisons []score.MetricComparison) []Regression {
	out := make([]Regression, 0)
	for _, comparison := range comparisons {
		if !comparison.Regressed {
			continue
		}
		out = append(out, NewRegressionFromMetricComparison(
			matchID,
			comparison,
			RegressionMinor,
			"candidate score is worse than baseline",
		))
	}
	return out
}
