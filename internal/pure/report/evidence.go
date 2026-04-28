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

// ProjectScoreEvidence projects a candidate report into the pure score evidence
// document used by artifact serialization and future objective layers.
func ProjectScoreEvidence(candidateReport CandidateReport) (score.ScoreEvidenceDocument, error) {
	if candidateReport.ID == "" {
		return score.ScoreEvidenceDocument{}, fmt.Errorf("%w: %w", ErrEvidenceProjectionFailed, ErrMissingReportID)
	}
	if err := candidateReport.Spec.Validate(); err != nil {
		return score.ScoreEvidenceDocument{}, fmt.Errorf("%w: %w: %v", ErrEvidenceProjectionFailed, ErrInvalidReportSpec, err)
	}

	metrics := make([]score.MetricEvidence, 0, len(candidateReport.Comparisons))
	for _, comparison := range candidateReport.Comparisons {
		metricEvidence, err := score.NewMetricEvidence(comparison.Metric, comparison.Baseline, comparison.Candidate)
		if err != nil {
			return score.ScoreEvidenceDocument{}, fmt.Errorf("%w: %w: %v", ErrEvidenceProjectionFailed, ErrInvalidMetricEvidence, err)
		}
		metrics = append(metrics, metricEvidence)
	}

	regressionDetails := make([]score.RegressionEvidence, 0, len(candidateReport.Regressions))
	regressionSummary := score.RegressionEvidenceSummary{
		Count: len(candidateReport.Regressions),
	}
	for _, regression := range candidateReport.Regressions {
		regressionDetails = append(regressionDetails, score.RegressionEvidence{
			TaskID:    regression.TaskID,
			Metric:    regression.Metric,
			Baseline:  regression.Baseline,
			Candidate: regression.Candidate,
			Delta:     regression.Delta,
			Severity:  string(regression.Severity),
			Reason:    regression.Reason,
		})
		switch regression.Severity {
		case RegressionMinor:
			regressionSummary.MinorCount++
		case RegressionBlocking:
			regressionSummary.SevereCount++
		}
	}

	return score.ScoreEvidenceDocument{
		SchemaVersion: score.EvidenceSchemaVersion,
		ReportID:      candidateReport.ID,
		Systems:       candidateReport.Spec.Systems,
		RunCounts: score.RoleCounts{
			Baseline:  len(candidateReport.Runs.Baseline),
			Candidate: len(candidateReport.Runs.Candidate),
		},
		FailureCounts: score.RoleCounts{
			Baseline:  len(candidateReport.Failures.Baseline),
			Candidate: len(candidateReport.Failures.Candidate),
		},
		LocalizationDistance: score.ExtractLocalizationDistance(metrics),
		Usage:                score.AggregateUsage(candidateReport.Runs.Candidate),
		BaselineUsage:        score.AggregateUsage(candidateReport.Runs.Baseline),
		Regressions:          regressionSummary,
		RegressionDetails:    regressionDetails,
		InvalidPredictions: score.InvalidPredictionEvidence{
			Known: false,
			Count: 0,
		},
		Metrics: metrics,
		PromotionDecision: score.PromotionDecisionEvidence{
			Decision: string(candidateReport.Decision.Decision),
			Reason:   candidateReport.Decision.Reason,
		},
	}, nil
}
