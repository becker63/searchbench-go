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

func loadEvidence(plan Plan) (pureoptimizer.NextChallengerEvidence, error) {
	evidence := pureoptimizer.NextChallengerEvidence{
		ParentRound:   plan.ParentBundle,
		IncludedKinds: append([]string(nil), plan.IncludedEvidence...),
		DeniedKinds:   append([]string(nil), plan.DeniedEvidence...),
	}

	source, err := os.ReadFile(plan.InputPolicy.Path)
	if err != nil {
		return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("read input policy source: %w", err)
	}
	evidence.InputPolicy = pureoptimizer.PolicySource{
		ArtifactID:  domain.ArtifactID(plan.InputPolicy.ArtifactID),
		Path:        filepath.ToSlash(plan.InputPolicy.Path),
		InterfaceID: plan.InputPolicy.InterfaceID,
		Source:      string(source),
	}

	needsReport := includesEvidence(plan, "report_summary") || includesEvidence(plan, "round_evidence")
	var roundReport purereport.RoundReport
	if needsReport {
		parentDir := string(plan.ParentBundle.BundlePath)
		reportPath := filepath.Join(parentDir, "round-report.json")
		if err := decodeJSONFile(reportPath, &roundReport); err != nil {
			return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("load parent report: %w", err)
		}
	}

	if includesEvidence(plan, "report_summary") {
		summary := summarizeReport(roundReport)
		evidence.ReportSummary = &summary
	}
	if includesEvidence(plan, "round_evidence") {
		roundEvidence, err := purereport.BuildRoundEvidence(roundReport)
		if err != nil {
			return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("project parent round evidence: %w", err)
		}
		evidence.RoundEvidence = &roundEvidence
	}
	if includesEvidence(plan, "objective_result") {
		var objectiveResult score.ObjectiveResult
		if err := decodeJSONFile(filepath.Join(string(plan.ParentBundle.BundlePath), "objective.json"), &objectiveResult); err != nil {
			return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("load objective result: %w", err)
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

func summarizeReport(roundReport purereport.RoundReport) pureoptimizer.ReportSummary {
	summary := pureoptimizer.ReportSummary{
		ReportID:       roundReport.ID,
		Decision:       string(roundReport.Decision.Decision),
		DecisionReason: roundReport.Decision.Reason,
		Comparisons:    make([]pureoptimizer.MetricSummary, 0, len(roundReport.Comparisons)),
	}

	if roundReport.Spec.Policies.Challenger.Policy != nil {
		summary.ChallengerPolicyID = roundReport.Spec.Policies.Challenger.Policy.ID
	}
	summary.ChallengerSystemID = roundReport.Spec.Policies.Challenger.ID

	challengerUsage := score.AggregateUsage(roundReport.Runs.Challenger)
	incumbentUsage := score.AggregateUsage(roundReport.Runs.Incumbent)
	summary.ChallengerUsage = challengerUsage
	summary.IncumbentUsage = incumbentUsage

	for _, comparison := range roundReport.Comparisons {
		summary.Comparisons = append(summary.Comparisons, pureoptimizer.MetricSummary{
			Metric:     string(comparison.Metric),
			Incumbent:  comparison.Incumbent,
			Challenger: comparison.Challenger,
			Delta:      comparison.Delta,
		})
	}
	return summary
}

func includesEvidence(plan Plan, kind string) bool {
	return slices.Contains(plan.IncludedEvidence, kind)
}
