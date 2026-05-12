package round

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	optimizereino "github.com/becker63/searchbench-go/internal/agents/optimizer/eino"
	optimizepolicy "github.com/becker63/searchbench-go/internal/agents/optimizer/policy"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	purereport "github.com/becker63/searchbench-go/internal/pure/report"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func materializeChallenger(ctx context.Context, resolved Resolved, input Input) (Resolved, error) {
	plan := resolved.Round
	if plan.ChallengerMaterialization.Mode != "generated" {
		return resolved, nil
	}
	if input.OptimizerModelFactory == nil {
		return Resolved{}, fmt.Errorf("round: OptimizerModelFactory is required for generated challenger materialization")
	}
	if plan.Optimizer == nil {
		return Resolved{}, fmt.Errorf("round: optimizer config is required for generated challenger materialization")
	}
	if plan.Lineage.Continues == "" {
		return Resolved{}, fmt.Errorf("round: generated challenger requires explicit continuation lineage")
	}

	parentBundle := domain.HostPath(plan.Lineage.Continues)
	continuation, err := bundlefs.LoadContinuation(parentBundle)
	if err != nil {
		return Resolved{}, err
	}
	if continuation.SurvivingCandidate.System.Policy == nil {
		return Resolved{}, fmt.Errorf("round: parent continuation has no surviving candidate policy to optimize")
	}

	optimizerModel, err := input.OptimizerModelFactory()
	if err != nil {
		return Resolved{}, err
	}

	evidence, err := buildGeneratedChallengerEvidence(parentBundle, continuation, plan)
	if err != nil {
		return Resolved{}, err
	}

	executor, err := optimizereino.NewOptimizer(optimizereino.OptimizerConfig{
		Model:            optimizerModel,
		ValidateProposal: optimizepolicy.Validate,
	})
	if err != nil {
		return Resolved{}, err
	}

	targetArtifactID := domain.PolicyID("next-challenger-" + plan.Round.ID)

	record := executor.Run(ctx, pureoptimizer.Spec{
		Target: pureoptimizer.NextChallengerTarget{
			InputArtifactID:  domain.ArtifactID(continuation.SurvivingCandidate.System.Policy.ID),
			OutputArtifactID: domain.ArtifactID(targetArtifactID),
			OutputName:       plan.ChallengerMaterialization.ArtifactName,
			InterfaceID:      plan.CandidateInterfaceID,
		},
		Agent:    plan.Optimizer.Agent,
		Evidence: evidence,
	})
	if record.Failure != nil {
		return Resolved{}, record.Failure
	}
	if record.Proposal == nil {
		return Resolved{}, fmt.Errorf("round: optimizer returned no generated challenger proposal")
	}

	policy := domain.NewPythonPolicy(domain.PolicyID(record.Proposal.ArtifactID), record.Proposal.Code, selectionPolicyV1DefaultSymbol)
	plan.Policies.Challenger.Policy = &policy
	plan.Output.ResolvedPolicyPaths.Challenger = filepath.ToSlash(filepath.Join(string(plan.Output.ExpectedBundlePath), "policies", record.Proposal.ArtifactName))
	resolved.Round = plan
	return resolved, nil
}

func buildGeneratedChallengerEvidence(
	parentBundle domain.HostPath,
	continuation pureround.Continuation,
	plan Plan,
) (pureoptimizer.NextChallengerEvidence, error) {
	inputPolicyPath := continuation.ResolveArtifactPath(parentBundle)
	source, err := os.ReadFile(inputPolicyPath)
	if err != nil {
		return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("read parent candidate policy: %w", err)
	}

	evidence := pureoptimizer.NextChallengerEvidence{
		ParentRound: pureoptimizer.ParentRoundRef{
			BundleID:   filepath.Base(string(parentBundle)),
			BundlePath: parentBundle,
		},
		IncludedKinds: append([]string(nil), plan.ChallengerMaterialization.IncludedEvidence...),
		DeniedKinds:   append([]string(nil), plan.ChallengerMaterialization.DeniedEvidence...),
		InputPolicy: pureoptimizer.PolicySource{
			ArtifactID:  domain.ArtifactID(continuation.SurvivingCandidate.System.Policy.ID),
			Path:        filepath.ToSlash(inputPolicyPath),
			InterfaceID: plan.CandidateInterfaceID,
			Source:      string(source),
		},
	}

	needsReport := includesGeneratedEvidence(plan, "report_summary") || includesGeneratedEvidence(plan, "round_evidence")
	var roundReport purereport.RoundReport
	if needsReport {
		if err := decodeJSONArtifact(filepath.Join(string(parentBundle), "round-report.json"), &roundReport); err != nil {
			return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("load parent report: %w", err)
		}
	}
	if includesGeneratedEvidence(plan, "report_summary") {
		summary := summarizeGeneratedReport(roundReport)
		evidence.ReportSummary = &summary
	}
	if includesGeneratedEvidence(plan, "round_evidence") {
		roundEvidence, err := purereport.BuildRoundEvidence(roundReport)
		if err != nil {
			return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("project parent round evidence: %w", err)
		}
		evidence.RoundEvidence = &roundEvidence
	}
	if includesGeneratedEvidence(plan, "objective_result") {
		var objectiveResult score.ObjectiveResult
		if err := decodeJSONArtifact(filepath.Join(string(parentBundle), "objective.json"), &objectiveResult); err != nil {
			return pureoptimizer.NextChallengerEvidence{}, fmt.Errorf("load objective result: %w", err)
		}
		evidence.ObjectiveResult = &objectiveResult
	}

	return evidence, nil
}

func includesGeneratedEvidence(plan Plan, kind string) bool {
	return slices.Contains(plan.ChallengerMaterialization.IncludedEvidence, kind)
}

func summarizeGeneratedReport(roundReport purereport.RoundReport) pureoptimizer.ReportSummary {
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
	summary.ChallengerUsage = score.AggregateUsage(roundReport.Runs.Challenger)
	summary.IncumbentUsage = score.AggregateUsage(roundReport.Runs.Incumbent)
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

func decodeJSONArtifact(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
