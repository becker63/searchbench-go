package report

import "github.com/becker63/searchbench-go/internal/pure/score"

// ScoreComparison compares one metric across incumbent and challenger.
//
// Positive/negative meaning depends on the metric. Higher-is-better or
// lower-is-better policy should live in promotion/regression logic, not here.
type ScoreComparison struct {
	Metric score.MetricName `json:"metric"`

	Incumbent  float64 `json:"incumbent"`
	Challenger float64 `json:"challenger"`
	Delta      float64 `json:"delta"`
}

// NewScoreComparison constructs a comparison using challenger - incumbent.
func NewScoreComparison(metric score.MetricName, incumbent, challenger float64) ScoreComparison {
	return ScoreComparison{
		Metric:     metric,
		Incumbent:  incumbent,
		Challenger: challenger,
		Delta:      challenger - incumbent,
	}
}

// NewScoreComparisonFromMetric converts a score-level metric comparison into
// its report-facing shape.
func NewScoreComparisonFromMetric(c score.MetricComparison) ScoreComparison {
	return ScoreComparison{
		Metric:     c.Metric,
		Incumbent:  c.Incumbent,
		Challenger: c.Challenger,
		Delta:      c.Delta,
	}
}
