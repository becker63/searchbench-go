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

// Regression records a metric/task-level reason the candidate got worse.
type Regression struct {
	TaskID domain.TaskID    `json:"task_id"`
	Metric score.MetricName `json:"metric"`

	Baseline  float64 `json:"baseline"`
	Candidate float64 `json:"candidate"`
	Delta     float64 `json:"delta"`

	Severity RegressionSeverity `json:"severity"`
	Reason   string             `json:"reason"`
}

// NewRegressionFromMetricComparison converts a score-level regression into a
// report-facing regression record.
func NewRegressionFromMetricComparison(
	taskID domain.TaskID,
	comparison score.MetricComparison,
	severity RegressionSeverity,
	reason string,
) Regression {
	return Regression{
		TaskID:    taskID,
		Metric:    comparison.Metric,
		Baseline:  comparison.Baseline,
		Candidate: comparison.Candidate,
		Delta:     comparison.Delta,
		Severity:  severity,
		Reason:    reason,
	}
}

// RegressionsForTask converts metric comparisons into report-level regressions
// for one task.
func RegressionsForTask(taskID domain.TaskID, comparisons []score.MetricComparison) []Regression {
	out := make([]Regression, 0)
	for _, comparison := range comparisons {
		if !comparison.Regressed {
			continue
		}
		out = append(out, NewRegressionFromMetricComparison(
			taskID,
			comparison,
			RegressionMinor,
			"candidate score is worse than baseline",
		))
	}
	return out
}
