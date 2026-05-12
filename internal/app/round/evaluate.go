package round

import (
	"context"
	"errors"
	"os"
	"time"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	text "github.com/becker63/searchbench-go/internal/adapters/report/text"
	scoring "github.com/becker63/searchbench-go/internal/adapters/scoring/pkl"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// runEvaluation executes the full manifest-driven evaluation pipeline. It is
// the convenience entry point used by tests and by the legacy single-call API;
// the production round flow runs the same helpers individually through the
// named phase functions in run.go.
func runEvaluation(ctx context.Context, request evaluationRequest) (Result, error) {
	plan, err := resolveEvaluation(ctx, request.Resolve)
	if err != nil {
		phase := PhaseResolvePlanFailed
		if errors.Is(err, config.ErrValidationFailed) {
			phase = PhaseValidateManifestFailed
		} else if errors.Is(err, os.ErrNotExist) {
			phase = PhaseResolvePlanFailed
		} else if request.Resolve.ManifestPath != "" {
			if absErr := normalizeManifestPathError(request.Resolve.ManifestPath, err); absErr != nil {
				return Result{}, &Error{Phase: PhaseLoadManifestFailed, Err: absErr}
			}
		}
		return Result{}, &Error{Phase: phase, Err: err}
	}
	return runEvaluationResolved(ctx, plan, request)
}

// runEvaluationResolved orchestrates the full evaluation pipeline once the
// plan is resolved. Each step delegates to the same helper that the named
// public phase functions call, so this wrapper and the phase functions stay
// in lockstep.
func runEvaluationResolved(ctx context.Context, plan Plan, request evaluationRequest) (Result, error) {
	resolved, err := materializeChallenger(ctx, Resolved{Round: plan}, Input{
		EvaluatorModelFactory: request.EvaluatorModelFactory,
		EvaluatorToolFactory:  request.EvaluatorToolFactory,
		OptimizerModelFactory: nil,
	})
	if err != nil {
		return Result{}, &Error{Phase: PhaseMaterializeChallengerFailed, Err: err}
	}
	plan = resolved.Round

	runCtx, cancel := withEvaluatorTimeout(ctx, plan)
	defer cancel()

	roundReport, executions, matchExec, err := runComparison(runCtx, plan, request)
	if err != nil {
		return Result{}, &Error{Phase: PhaseComparisonFailed, Err: err}
	}

	evidence, err := projectRoundEvidence(plan, roundReport, matchExec)
	if err != nil {
		return Result{}, &Error{Phase: PhaseRoundEvidenceFailed, Err: err}
	}

	objective, err := evaluateObjectiveForPlan(ctx, plan, evidence, request)
	if err != nil {
		return Result{}, err
	}

	bundleRef, err := writeRoundBundle(ctx, plan, request, roundReport, evidence, objective)
	if err != nil {
		return Result{}, err
	}

	return Result{
		ManifestPath:        plan.ManifestPath,
		Bundle:              bundleRef,
		ReportID:            roundReport.ID,
		RoundReport:         roundReport,
		RoundEvidence:       evidence,
		ObjectiveResult:     objective,
		EvaluatorExecutions: executions,
		MatchExecutions:     matchExec,
	}, nil
}

func withEvaluatorTimeout(ctx context.Context, plan Plan) (context.Context, context.CancelFunc) {
	timeout := timeoutFromSeconds(plan.Evaluator.Bounds.TimeoutSeconds)
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}

func buildEvidenceFromReport(plan Plan, roundReport report.RoundReport) (score.RoundEvidenceDocument, error) {
	evidence, err := report.BuildRoundEvidence(roundReport)
	if err != nil {
		return score.RoundEvidenceDocument{}, err
	}
	evidence.GameID = plan.Game.ID
	evidence.RoundID = plan.Round.ID
	if err := evidence.Validate(); err != nil {
		return score.RoundEvidenceDocument{}, err
	}
	return evidence, nil
}

func projectRoundEvidence(plan Plan, roundReport report.RoundReport, matchExec []report.MatchExecutionRecord) (score.RoundEvidenceDocument, error) {
	evidence, err := buildEvidenceFromReport(plan, roundReport)
	if err != nil {
		return score.RoundEvidenceDocument{}, err
	}
	enrichLocalizationEvidence(&evidence, matchExec)
	return evidence, nil
}

func evaluateObjectiveForPlan(ctx context.Context, plan Plan, evidence score.RoundEvidenceDocument, request evaluationRequest) (*score.ObjectiveResult, error) {
	evidenceInput, err := materializeRoundEvidence(plan, evidence)
	if err != nil {
		return nil, &Error{Phase: PhaseEvidencePKLFailed, Err: err}
	}
	defer evidenceInput.Cleanup()

	objectiveResult, err := scoring.Evaluate(ctx, scoring.Request{
		ScoringPath:         plan.Scoring.ObjectivePath,
		CurrentRef:          evidenceInput.CurrentRef,
		CurrentEvidencePath: evidenceInput.CurrentEvidencePath,
		ParentRef:           evidenceInput.ParentRef,
		ParentEvidencePath:  evidenceInput.ParentEvidencePath,
		PklCommand:          request.PklCommand,
	})
	if err != nil {
		return nil, &Error{Phase: PhaseObjectiveFailed, Err: err}
	}
	return &objectiveResult, nil
}

func writeRoundBundle(
	ctx context.Context,
	plan Plan,
	request evaluationRequest,
	roundReport report.RoundReport,
	evidence score.RoundEvidenceDocument,
	objective *score.ObjectiveResult,
) (bundlefs.BundleRef, error) {
	rendered, err := renderReport(plan, request, roundReport)
	if err != nil {
		return bundlefs.BundleRef{}, &Error{Phase: PhaseRenderReportFailed, Err: err}
	}
	additionalFiles, policyPaths := roundBundleArtifacts(plan)
	continuation := buildContinuation(plan, roundReport, objective, policyPaths)
	continuationPKL, err := buildContinuationPKLInput(plan)
	if err != nil {
		return bundlefs.BundleRef{}, &Error{Phase: PhaseBundleWriteFailed, Err: err}
	}
	bundleRef, err := bundlefs.WriteBundle(ctx, bundlefs.RoundBundleInput{
		RootPath:        plan.Output.BundleWriterRoot,
		BundleID:        plan.Bundle.ID,
		ResolvedInput:   plan,
		RoundReport:     roundReport,
		RoundEvidence:   evidence,
		ObjectiveResult: objective,
		RenderedReport:  rendered,
		Continuation:    continuation,
		ContinuationPKL: continuationPKL,
		AdditionalFiles: additionalFiles,
		CreatedAt:       plan.CreatedAt,
	})
	if err != nil {
		return bundlefs.BundleRef{}, &Error{Phase: PhaseBundleWriteFailed, Err: err}
	}
	return bundleRef, nil
}

func renderReport(plan Plan, request evaluationRequest, roundReport report.RoundReport) (*bundlefs.RenderedReport, error) {
	if !plan.Output.RenderHumanReport || request.DisableRenderReport {
		return nil, nil
	}
	content := text.RenderRoundReport(roundReport, text.Options{
		Color: false,
		Width: 100,
	})
	return &bundlefs.RenderedReport{
		FileName: "round-report.txt",
		Content:  content + "\n",
	}, nil
}

func normalizeManifestPathError(manifestPath string, err error) error {
	if manifestPath == "" {
		return nil
	}
	if _, statErr := os.Stat(manifestPath); statErr != nil {
		return statErr
	}
	return nil
}

func timeoutFromSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
