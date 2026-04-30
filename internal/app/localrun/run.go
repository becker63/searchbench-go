package localrun

import (
	"context"
	"errors"
	"os"
	"time"

	artifact "github.com/becker63/searchbench-go/internal/adapters/artifact/fsbundle"
	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	scoring "github.com/becker63/searchbench-go/internal/adapters/scoring/pkl"
	appExperiment "github.com/becker63/searchbench-go/internal/app/experiment"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/surface/console"
)

// Run executes the smallest manifest-driven local fake SearchBench-Go path and
// writes one immutable bundle.
func Run(ctx context.Context, request Request) (Result, error) {
	plan, err := appExperiment.Resolve(ctx, request)
	if err != nil {
		phase := PhaseResolvePlanFailed
		if errors.Is(err, appExperiment.ErrUnsupportedMode) {
			phase = PhaseUnsupportedMode
		} else if errors.Is(err, config.ErrValidationFailed) {
			phase = PhaseValidateManifestFailed
		} else if errors.Is(err, os.ErrNotExist) {
			phase = PhaseResolvePlanFailed
		} else if request.ManifestPath != "" {
			if absErr := normalizeManifestPathError(request.ManifestPath, err); absErr != nil {
				return Result{}, &Error{Phase: PhaseLoadManifestFailed, Err: absErr}
			}
		}
		return Result{}, &Error{Phase: phase, Err: err}
	}

	runCtx := ctx
	cancel := func() {}
	timeout := timeoutFromSeconds(plan.Evaluator.Bounds.TimeoutSeconds)
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	candidateReport, err := runFakeComparison(runCtx, plan)
	if err != nil {
		return Result{}, &Error{Phase: PhaseFakeComparisonFailed, Err: err}
	}

	scoreEvidence, err := report.ProjectScoreEvidence(candidateReport)
	if err != nil {
		return Result{}, &Error{Phase: PhaseScoreEvidenceFailed, Err: err}
	}
	if err := scoreEvidence.Validate(); err != nil {
		return Result{}, &Error{Phase: PhaseScoreEvidenceFailed, Err: err}
	}

	evidenceInput, err := materializeScoreEvidence(plan, scoreEvidence)
	if err != nil {
		return Result{}, &Error{Phase: PhaseScorePKLFailed, Err: err}
	}
	defer evidenceInput.Cleanup()

	objectiveResult, err := scoring.Evaluate(ctx, scoring.Request{
		ScoringPath:      plan.Scoring.ObjectivePath,
		CurrentRef:       evidenceInput.CurrentRef,
		CurrentScorePath: evidenceInput.CurrentScorePath,
		ParentRef:        evidenceInput.ParentRef,
		ParentScorePath:  evidenceInput.ParentScorePath,
		PklCommand:       request.PklCommand,
	})
	if err != nil {
		return Result{}, &Error{Phase: PhaseObjectiveFailed, Err: err}
	}

	renderedReport, err := renderReport(plan, candidateReport)
	if err != nil {
		return Result{}, &Error{Phase: PhaseRenderReportFailed, Err: err}
	}

	bundleRef, err := artifact.WriteBundle(ctx, artifact.BundleRequest{
		RootPath:        plan.Output.BundleWriterRoot,
		BundleID:        plan.Bundle.ID,
		ResolvedInput:   bundleResolvedInput(plan),
		CandidateReport: candidateReport,
		ScoreEvidence:   scoreEvidence,
		ObjectiveResult: &objectiveResult,
		RenderedReport:  renderedReport,
		CreatedAt:       plan.CreatedAt,
	})
	if err != nil {
		return Result{}, &Error{Phase: PhaseBundleWriteFailed, Err: err}
	}

	return Result{
		ManifestPath:    plan.ManifestPath,
		Bundle:          bundleRef,
		ReportID:        candidateReport.ID,
		CandidateReport: candidateReport,
		ScoreEvidence:   scoreEvidence,
		ObjectiveResult: &objectiveResult,
	}, nil
}

func renderReport(plan appExperiment.ResolvedExperiment, candidateReport report.CandidateReport) (*artifact.RenderedReport, error) {
	if !plan.Output.RenderHumanReport {
		return nil, nil
	}
	content := console.RenderCandidateReport(candidateReport, console.Options{
		Color: false,
		Width: 100,
	})
	return &artifact.RenderedReport{
		FileName: "report.txt",
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
