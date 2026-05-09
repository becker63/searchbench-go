package report

import (
	"errors"
	"fmt"

	"github.com/becker63/searchbench-go/internal/pure/score"
)

var (
	ErrEvidenceProjectionFailed = errors.New("report: evidence projection failed")
	ErrMissingReportID          = errors.New("report: missing report id")
	ErrInvalidReportSpec        = errors.New("report: invalid report spec")
	ErrInvalidMetricEvidence    = errors.New("report: invalid metric evidence")
)

// BuildRoundEvidence projects a round report into the pure round evidence
// document used by artifact serialization and objective layers.
func BuildRoundEvidence(roundReport RoundReport) (score.RoundEvidenceDocument, error) {
	if roundReport.ID == "" {
		return score.RoundEvidenceDocument{}, fmt.Errorf("%w: %w", ErrEvidenceProjectionFailed, ErrMissingReportID)
	}
	if err := roundReport.Spec.Validate(); err != nil {
		return score.RoundEvidenceDocument{}, fmt.Errorf("%w: %w: %v", ErrEvidenceProjectionFailed, ErrInvalidReportSpec, err)
	}

	metrics := make([]score.MetricEvidence, 0, len(roundReport.Comparisons))
	for _, comparison := range roundReport.Comparisons {
		metricEvidence, err := score.NewMetricEvidence(comparison.Metric, comparison.Baseline, comparison.Candidate)
		if err != nil {
			return score.RoundEvidenceDocument{}, fmt.Errorf("%w: %w: %v", ErrEvidenceProjectionFailed, ErrInvalidMetricEvidence, err)
		}
		metrics = append(metrics, metricEvidence)
	}

	regressionDetails := make([]score.RegressionEvidence, 0, len(roundReport.Regressions))
	regressionSummary := score.RegressionEvidenceSummary{
		Count: len(roundReport.Regressions),
	}
	for _, regression := range roundReport.Regressions {
		regressionDetails = append(regressionDetails, score.RegressionEvidence{
			MatchID:    regression.MatchID,
			Metric:     regression.Metric,
			Incumbent:  regression.Baseline,
			Challenger: regression.Candidate,
			Delta:      regression.Delta,
			Severity:   string(regression.Severity),
			Reason:     regression.Reason,
		})
		switch regression.Severity {
		case RegressionMinor:
			regressionSummary.MinorCount++
		case RegressionBlocking:
			regressionSummary.SevereCount++
		}
	}

	return score.RoundEvidenceDocument{
		SchemaVersion: score.EvidenceSchemaVersion,
		ReportID:      roundReport.ID,
		Policies:      roundReport.Spec.Systems,
		MatchCounts: score.MatchCounts{
			Total: roundReport.Spec.Matches.Len(),
		},
		ExecutionCounts: score.RoleCounts{
			Incumbent:  len(roundReport.Runs.Incumbent),
			Challenger: len(roundReport.Runs.Challenger),
		},
		FailureCounts: score.RoleCounts{
			Incumbent:  len(roundReport.Failures.Incumbent),
			Challenger: len(roundReport.Failures.Challenger),
		},
		LocalizationDistance: score.ExtractLocalizationDistance(metrics),
		ChallengerUsage:      score.AggregateUsage(roundReport.Runs.Challenger),
		IncumbentUsage:       score.AggregateUsage(roundReport.Runs.Incumbent),
		Regressions:          regressionSummary,
		RegressionDetails:    regressionDetails,
		InvalidPredictions: score.InvalidPredictionEvidence{
			Known: false,
			Count: 0,
		},
		Metrics: metrics,
		Decision: score.DecisionEvidence{
			Decision: string(roundReport.Decision.Decision),
			Reason:   roundReport.Decision.Reason,
		},
	}, nil
}

// ProjectScoreEvidence is a transitional wrapper for BuildRoundEvidence.
//
// TODO(issue-32): remove after callers use BuildRoundEvidence directly.
func ProjectScoreEvidence(roundReport RoundReport) (score.RoundEvidenceDocument, error) {
	return BuildRoundEvidence(roundReport)
}
