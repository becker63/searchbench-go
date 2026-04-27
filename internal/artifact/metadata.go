package artifact

import (
	"encoding/hex"
	"path/filepath"
	"time"

	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/score"
)

func buildScoreEvidence(candidateReport report.CandidateReport) ScoreEvidence {
	metrics := make([]MetricEvidence, 0, len(candidateReport.Comparisons))
	for _, comparison := range candidateReport.Comparisons {
		direction, _ := score.DirectionForMetric(comparison.Metric)
		improved := score.Improved(direction, comparison.Baseline, comparison.Candidate)
		regressed := comparison.Baseline != comparison.Candidate && !improved
		metrics = append(metrics, MetricEvidence{
			Metric:    comparison.Metric,
			Direction: direction,
			Baseline:  comparison.Baseline,
			Candidate: comparison.Candidate,
			Delta:     comparison.Delta,
			Improved:  improved,
			Regressed: regressed,
		})
	}

	return ScoreEvidence{
		PromotionDecision: candidateReport.Decision,
		RunCounts: RoleCounts{
			Baseline:  len(candidateReport.Runs.Baseline),
			Candidate: len(candidateReport.Runs.Candidate),
		},
		FailureCounts: RoleCounts{
			Baseline:  len(candidateReport.Failures.Baseline),
			Candidate: len(candidateReport.Failures.Candidate),
		},
		Metrics:     metrics,
		Regressions: append([]report.Regression(nil), candidateReport.Regressions...),
	}
}

func buildMetadata(bundleID string, createdAt time.Time, files []BundleFile) BundleMetadata {
	return BundleMetadata{
		SchemaVersion: schemaVersion,
		BundleID:      bundleID,
		CreatedAt:     createdAt.UTC(),
		Files:         append([]BundleFile(nil), files...),
	}
}

func fileRecord(kind string, path string, mediaType string, sha []byte) BundleFile {
	file := BundleFile{
		Kind:      kind,
		Path:      filepath.ToSlash(path),
		MediaType: mediaType,
	}
	if len(sha) > 0 {
		file.SHA256 = hex.EncodeToString(sha)
	}
	return file
}
