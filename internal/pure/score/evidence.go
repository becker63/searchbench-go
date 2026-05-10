package score

import (
	"errors"
	"fmt"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

const EvidenceSchemaVersion = "searchbench.round_evidence.v1"

var (
	ErrUnsupportedMetricDirection   = errors.New("score: unsupported metric direction")
	ErrInvalidRoundEvidence         = errors.New("score: invalid round evidence")
	ErrMissingEvidenceSchemaVersion = errors.New("score: missing round evidence schema version")
	ErrMissingEvidenceReportID      = errors.New("score: missing round evidence report id")
)

// RoundEvidenceDocument is the objective-ready durable evidence view derived
// from a round report.
//
// It is intentionally field-addressable so future objective layers can read
// durable facts without parsing report tables or artifact package internals.
type RoundEvidenceDocument struct {
	SchemaVersion        string                        `json:"schema_version"`
	GameID               string                        `json:"game_id,omitempty"`
	RoundID              string                        `json:"round_id,omitempty"`
	ReportID             domain.ReportID               `json:"report_id"`
	Policies             domain.Pair[domain.SystemRef] `json:"policies"`
	MatchCounts          MatchCounts                   `json:"match_counts"`
	ExecutionCounts      RoleCounts                    `json:"execution_counts"`
	FailureCounts        RoleCounts                    `json:"failure_counts"`
	LocalizationDistance LocalizationDistanceEvidence  `json:"localization_distance"`
	ChallengerUsage      UsageEvidence                 `json:"challenger_usage"`
	IncumbentUsage       UsageEvidence                 `json:"incumbent_usage"`
	Regressions          RegressionEvidenceSummary     `json:"regressions"`
	RegressionDetails    []RegressionEvidence          `json:"regression_details,omitempty"`
	InvalidPredictions   InvalidPredictionEvidence     `json:"invalid_predictions"`
	Metrics              []MetricEvidence              `json:"metrics"`
	Decision             DecisionEvidence              `json:"decision"`
}

// MatchCounts records how many matches were included in the round evidence.
type MatchCounts struct {
	Total int `json:"total"`
}

// RoleCounts records incumbent/challenger counts in stable role order.
type RoleCounts struct {
	Incumbent  int `json:"incumbent"`
	Challenger int `json:"challenger"`
}

// MetricEvidence is the pure metric-comparison evidence shape.
type MetricEvidence struct {
	Metric     MetricName `json:"metric"`
	Direction  Direction  `json:"direction"`
	Incumbent  float64    `json:"incumbent"`
	Challenger float64    `json:"challenger"`
	Delta      float64    `json:"delta"`
	Improved   bool       `json:"improved"`
	Regressed  bool       `json:"regressed"`
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
	MatchID    domain.MatchID `json:"match_id"`
	Metric     MetricName     `json:"metric"`
	Incumbent  float64        `json:"incumbent"`
	Challenger float64        `json:"challenger"`
	Delta      float64        `json:"delta"`
	Severity   string         `json:"severity"`
	Reason     string         `json:"reason"`
}

// InvalidPredictionEvidence is explicit about whether invalid-prediction
// counts are known in the current report substrate.
type InvalidPredictionEvidence struct {
	Known bool `json:"known"`
	Count int  `json:"count"`
}

// DecisionEvidence is the stable objective-facing decision summary.
type DecisionEvidence struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// Validate checks that the evidence document has the minimum structure needed
// for persistence and future objective use.
func (d RoundEvidenceDocument) Validate() error {
	if strings.TrimSpace(d.SchemaVersion) == "" {
		return fmt.Errorf("%w: %w", ErrInvalidRoundEvidence, ErrMissingEvidenceSchemaVersion)
	}
	if strings.TrimSpace(d.ReportID.String()) == "" {
		return fmt.Errorf("%w: %w", ErrInvalidRoundEvidence, ErrMissingEvidenceReportID)
	}
	return nil
}

// NewMetricEvidence constructs evidence for one metric comparison using the
// canonical metric direction.
func NewMetricEvidence(metric MetricName, incumbent, challenger float64) (MetricEvidence, error) {
	direction, ok := DirectionForMetric(metric)
	if !ok {
		return MetricEvidence{}, ErrUnsupportedMetricDirection
	}
	improved := Improved(direction, incumbent, challenger)
	return MetricEvidence{
		Metric:     metric,
		Direction:  direction,
		Incumbent:  incumbent,
		Challenger: challenger,
		Delta:      challenger - incumbent,
		Improved:   improved,
		Regressed:  incumbent != challenger && !improved,
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
