package optimizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	purereport "github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func loadEvidence(plan Plan) (pureoptimizer.Evidence, error) {
	evidence := pureoptimizer.Evidence{
		ParentRun:     plan.ParentBundle,
		IncludedKinds: append([]string(nil), plan.IncludedEvidence...),
		DeniedKinds:   append([]string(nil), plan.DeniedEvidence...),
	}

	source, err := os.ReadFile(plan.InputPolicy.Path)
	if err != nil {
		return pureoptimizer.Evidence{}, fmt.Errorf("read input policy source: %w", err)
	}
	evidence.InputPolicy = pureoptimizer.PolicySource{
		ArtifactID:  domain.ArtifactID(plan.InputPolicy.ArtifactID),
		Path:        filepath.ToSlash(plan.InputPolicy.Path),
		InterfaceID: plan.InputPolicy.InterfaceID,
		Source:      string(source),
	}

	needsReport := includesEvidence(plan, "report_summary") || includesEvidence(plan, "score_evidence")
	var candidateReport purereport.CandidateReport
	if needsReport {
		if err := decodeJSONFile(filepath.Join(string(plan.ParentBundle.BundlePath), "report.json"), &candidateReport); err != nil {
			return pureoptimizer.Evidence{}, fmt.Errorf("load parent report: %w", err)
		}
	}

	if includesEvidence(plan, "report_summary") {
		summary := summarizeReport(candidateReport)
		evidence.ReportSummary = &summary
	}
	if includesEvidence(plan, "score_evidence") {
		scoreEvidence, err := purereport.ProjectScoreEvidence(candidateReport)
		if err != nil {
			return pureoptimizer.Evidence{}, fmt.Errorf("project parent score evidence: %w", err)
		}
		evidence.ScoreEvidence = &scoreEvidence
	}
	if includesEvidence(plan, "objective_result") {
		var objectiveResult score.ObjectiveResult
		if err := decodeJSONFile(filepath.Join(string(plan.ParentBundle.BundlePath), "objective.json"), &objectiveResult); err != nil {
			return pureoptimizer.Evidence{}, fmt.Errorf("load objective result: %w", err)
		}
		evidence.ObjectiveResult = &objectiveResult
	}

	return evidence, nil
}

func decodeJSONFile(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}

func summarizeReport(candidateReport purereport.CandidateReport) pureoptimizer.ReportSummary {
	summary := pureoptimizer.ReportSummary{
		ReportID:       candidateReport.ID,
		Decision:       string(candidateReport.Decision.Decision),
		DecisionReason: candidateReport.Decision.Reason,
		Comparisons:    make([]pureoptimizer.MetricSummary, 0, len(candidateReport.Comparisons)),
	}

	if candidateReport.Spec.Systems.Candidate.Policy != nil {
		summary.CandidatePolicyID = candidateReport.Spec.Systems.Candidate.Policy.ID
	}
	summary.CandidateSystemID = candidateReport.Spec.Systems.Candidate.ID

	candidateUsage := score.AggregateUsage(candidateReport.Runs.Candidate)
	baselineUsage := score.AggregateUsage(candidateReport.Runs.Baseline)
	summary.CandidateUsage = candidateUsage
	summary.BaselineUsage = baselineUsage

	for _, comparison := range candidateReport.Comparisons {
		summary.Comparisons = append(summary.Comparisons, pureoptimizer.MetricSummary{
			Metric:    string(comparison.Metric),
			Baseline:  comparison.Baseline,
			Candidate: comparison.Candidate,
			Delta:     comparison.Delta,
		})
	}
	return summary
}

func includesEvidence(plan Plan, kind string) bool {
	return slices.Contains(plan.IncludedEvidence, kind)
}
