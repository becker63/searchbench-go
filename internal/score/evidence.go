package score

import (
	"errors"

	"github.com/becker63/searchbench-go/internal/domain"
)

const EvidenceSchemaVersion = "searchbench.score_evidence.v1"

var ErrUnsupportedMetricDirection = errors.New("score: unsupported metric direction")

// ScoreEvidenceDocument is the objective-ready raw evidence view derived from a
// candidate report.
//
// It is intentionally field-addressable so future objective layers can read
// durable facts without parsing report tables or artifact package internals.
type ScoreEvidenceDocument struct {
	SchemaVersion        string                        `json:"schema_version"`
	ReportID             domain.ReportID               `json:"report_id"`
	Systems              domain.Pair[domain.SystemRef] `json:"systems"`
	RunCounts            RoleCounts                    `json:"run_counts"`
	FailureCounts        RoleCounts                    `json:"failure_counts"`
	LocalizationDistance LocalizationDistanceEvidence  `json:"localization_distance"`
	Usage                UsageEvidence                 `json:"usage"`
	BaselineUsage        UsageEvidence                 `json:"baseline_usage"`
	Regressions          RegressionEvidenceSummary     `json:"regressions"`
	RegressionDetails    []RegressionEvidence          `json:"regression_details,omitempty"`
	InvalidPredictions   InvalidPredictionEvidence     `json:"invalid_predictions"`
	Metrics              []MetricEvidence              `json:"metrics"`
	PromotionDecision    PromotionDecisionEvidence     `json:"promotion_decision"`
}

// RoleCounts records baseline/candidate counts in stable role order.
type RoleCounts struct {
	Baseline  int `json:"baseline"`
	Candidate int `json:"candidate"`
}

// MetricEvidence is the pure metric-comparison evidence shape.
type MetricEvidence struct {
	Metric    MetricName `json:"metric"`
	Direction Direction  `json:"direction"`
	Baseline  float64    `json:"baseline"`
	Candidate float64    `json:"candidate"`
	Delta     float64    `json:"delta"`
	Improved  bool       `json:"improved"`
	Regressed bool       `json:"regressed"`
}

// LocalizationDistanceEvidence exposes existing localization-oriented metric
// comparisons without inventing a new reducer.
type LocalizationDistanceEvidence struct {
	GoldHop  *MetricEvidence `json:"gold_hop,omitempty"`
	IssueHop *MetricEvidence `json:"issue_hop,omitempty"`
}

// UsageEvidence aggregates usage summaries while remaining honest about
// whether usage was actually measured.
type UsageEvidence struct {
	Available    bool              `json:"available"`
	MeasuredRuns int               `json:"measured_runs"`
	InputTokens  domain.TokenCount `json:"input_tokens"`
	OutputTokens domain.TokenCount `json:"output_tokens"`
	TotalTokens  domain.TokenCount `json:"total_tokens"`
	CostUSD      float64           `json:"cost_usd"`
}

// RegressionEvidenceSummary groups regressions into stable summary counts.
type RegressionEvidenceSummary struct {
	Count       int `json:"count"`
	MinorCount  int `json:"minor_count"`
	SevereCount int `json:"severe_count"`
}

// RegressionEvidence preserves report-level regression detail in a score-owned
// evidence shape.
type RegressionEvidence struct {
	TaskID    domain.TaskID `json:"task_id"`
	Metric    MetricName    `json:"metric"`
	Baseline  float64       `json:"baseline"`
	Candidate float64       `json:"candidate"`
	Delta     float64       `json:"delta"`
	Severity  string        `json:"severity"`
	Reason    string        `json:"reason"`
}

// InvalidPredictionEvidence is explicit about whether invalid-prediction
// counts are known in the current report substrate.
type InvalidPredictionEvidence struct {
	Known bool `json:"known"`
	Count int  `json:"count"`
}

// PromotionDecisionEvidence is the stable objective-facing decision summary.
type PromotionDecisionEvidence struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// NewMetricEvidence constructs evidence for one metric comparison using the
// canonical metric direction.
func NewMetricEvidence(metric MetricName, baseline, candidate float64) (MetricEvidence, error) {
	direction, ok := DirectionForMetric(metric)
	if !ok {
		return MetricEvidence{}, ErrUnsupportedMetricDirection
	}
	improved := Improved(direction, baseline, candidate)
	return MetricEvidence{
		Metric:    metric,
		Direction: direction,
		Baseline:  baseline,
		Candidate: candidate,
		Delta:     candidate - baseline,
		Improved:  improved,
		Regressed: baseline != candidate && !improved,
	}, nil
}

// AggregateUsage totals measured usage across scored runs.
func AggregateUsage(runs []ScoredRun) UsageEvidence {
	var usage UsageEvidence
	for _, run := range runs {
		summary := run.Execution.Usage
		if summary.Empty() {
			continue
		}
		usage.Available = true
		usage.MeasuredRuns++
		usage.InputTokens += summary.InputTokens
		usage.OutputTokens += summary.OutputTokens
		usage.TotalTokens += summary.TotalTokens
		usage.CostUSD += summary.CostUSD
	}
	return usage
}

// ExtractLocalizationDistance projects localization-oriented metrics into
// named fields so future objective layers do not need to scan metric arrays.
func ExtractLocalizationDistance(metrics []MetricEvidence) LocalizationDistanceEvidence {
	var out LocalizationDistanceEvidence
	for i := range metrics {
		metric := metrics[i]
		switch metric.Metric {
		case MetricGoldHop:
			copied := metric
			out.GoldHop = &copied
		case MetricIssueHop:
			copied := metric
			out.IssueHop = &copied
		}
	}
	return out
}
