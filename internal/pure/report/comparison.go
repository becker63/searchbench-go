package report

import "github.com/becker63/searchbench-go/internal/pure/score"

// ScoreComparison compares one metric across baseline and candidate.
//
// Positive/negative meaning depends on the metric. Higher-is-better or
// lower-is-better policy should live in promotion/regression logic, not here.
type ScoreComparison struct {
	Metric score.MetricName `json:"metric"`

	Baseline  float64 `json:"baseline"`
	Candidate float64 `json:"candidate"`
	Delta     float64 `json:"delta"`
}

// NewScoreComparison constructs a comparison using candidate - baseline.
func NewScoreComparison(metric score.MetricName, baseline, candidate float64) ScoreComparison {
	return ScoreComparison{
		Metric:    metric,
		Baseline:  baseline,
		Candidate: candidate,
		Delta:     candidate - baseline,
	}
}

// NewScoreComparisonFromMetric converts a score-level metric comparison into
// its report-facing shape.
func NewScoreComparisonFromMetric(c score.MetricComparison) ScoreComparison {
	return ScoreComparison{
		Metric:    c.Metric,
		Baseline:  c.Baseline,
		Candidate: c.Candidate,
		Delta:     c.Delta,
	}
}
