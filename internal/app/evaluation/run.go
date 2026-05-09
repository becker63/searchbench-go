package evaluation

import (
	"context"
	"errors"
	"os"
	"time"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	scoring "github.com/becker63/searchbench-go/internal/adapters/scoring/pkl"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/surface/console"
)

// Run executes the smallest manifest-driven local fake SearchBench-Go path and
// writes one immutable bundle.
func Run(ctx context.Context, request Request) (Result, error) {
	plan, err := Resolve(ctx, request.Resolve)
	if err != nil {
		phase := PhaseResolvePlanFailed
		if errors.Is(err, ErrUnsupportedMode) {
			phase = PhaseUnsupportedMode
		} else if errors.Is(err, config.ErrValidationFailed) {
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
	return RunResolved(ctx, plan, request)
}

// RunResolved executes one already-resolved local evaluation plan.
func RunResolved(ctx context.Context, plan Plan, request Request) (Result, error) {
	runCtx := ctx
	cancel := func() {}
	timeout := timeoutFromSeconds(plan.Evaluator.Bounds.TimeoutSeconds)
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	roundReport, evaluatorExecutions, err := runComparison(runCtx, plan, request)
	if err != nil {
		return Result{}, &Error{Phase: PhaseComparisonFailed, Err: err}
	}

	roundEvidence, err := report.BuildRoundEvidence(roundReport)
	if err != nil {
		return Result{}, &Error{Phase: PhaseScoreEvidenceFailed, Err: err}
	}
	roundEvidence.GameID = plan.Game.ID
	roundEvidence.RoundID = plan.Round.ID
	if err := roundEvidence.Validate(); err != nil {
		return Result{}, &Error{Phase: PhaseScoreEvidenceFailed, Err: err}
	}

	evidenceInput, err := materializeScoreEvidence(plan, roundEvidence)
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

	renderedReport, err := renderReport(plan, request, roundReport)
	if err != nil {
		return Result{}, &Error{Phase: PhaseRenderReportFailed, Err: err}
	}

	bundleRef, err := bundlefs.WriteBundle(ctx, bundlefs.RoundBundleInput{
		RootPath:        plan.Output.BundleWriterRoot,
		BundleID:        plan.Bundle.ID,
		ResolvedInput:   plan,
		RoundReport:     roundReport,
		RoundEvidence:   roundEvidence,
		ObjectiveResult: &objectiveResult,
		RenderedReport:  renderedReport,
		CreatedAt:       plan.CreatedAt,
	})
	if err != nil {
		return Result{}, &Error{Phase: PhaseBundleWriteFailed, Err: err}
	}

	return Result{
		ManifestPath:        plan.ManifestPath,
		Bundle:              bundleRef,
		ReportID:            roundReport.ID,
		RoundReport:         roundReport,
		RoundEvidence:       roundEvidence,
		ObjectiveResult:     &objectiveResult,
		EvaluatorExecutions: evaluatorExecutions,
	}, nil
}

func renderReport(plan Plan, request Request, roundReport report.RoundReport) (*bundlefs.RenderedReport, error) {
	if !plan.Output.RenderHumanReport || request.DisableRenderReport {
		return nil, nil
	}
	content := console.RenderRoundReport(roundReport, console.Options{
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
