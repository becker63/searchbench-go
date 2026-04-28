package localrun

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/becker63/searchbench-go/internal/adapters/artifact/fsbundle"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	"github.com/becker63/searchbench-go/internal/adapters/scoring/pkl"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
	"github.com/becker63/searchbench-go/internal/surface/console"
)

// Run executes the smallest manifest-driven local fake SearchBench-Go path and
// writes one immutable bundle.
func Run(ctx context.Context, request Request) (Result, error) {
	request = normalizeRequest(request)
	manifestPath, err := filepath.Abs(request.ManifestPath)
	if err != nil {
		return Result{}, &Error{Phase: PhaseLoadManifestFailed, Err: fmt.Errorf("resolve manifest path: %w", err)}
	}

	experiment, err := config.ResolveFromPath(ctx, manifestPath)
	if err != nil {
		return Result{}, &Error{Phase: PhaseLoadManifestFailed, Err: err}
	}
	if err := config.Validate(experiment); err != nil {
		return Result{}, &Error{Phase: PhaseValidateManifestFailed, Err: err}
	}

	projected, err := projectFakeRun(manifestPath, experiment, request)
	if err != nil {
		phase := PhaseProjectFakePlanFailed
		if errors.Is(err, errUnsupportedMode) {
			phase = PhaseUnsupportedMode
		}
		return Result{}, &Error{Phase: phase, Err: err}
	}

	candidateReport, err := runFakeComparison(ctx, projected)
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

	scoreInputPath, currentRef, cleanup, err := prepareScorePKL(projected, scoreEvidence)
	if err != nil {
		return Result{}, &Error{Phase: PhaseScorePKLFailed, Err: err}
	}
	defer cleanup()

	objectiveResult, err := scoring.Evaluate(ctx, scoring.Request{
		ScoringPath:      projected.objectivePath,
		CurrentRef:       currentRef,
		CurrentScorePath: scoreInputPath,
		PklCommand:       request.PklCommand,
	})
	if err != nil {
		return Result{}, &Error{Phase: PhaseObjectiveFailed, Err: err}
	}

	renderedReport, err := renderReport(projected, candidateReport)
	if err != nil {
		return Result{}, &Error{Phase: PhaseRenderReportFailed, Err: err}
	}

	bundleRef, err := artifact.WriteBundle(ctx, artifact.BundleRequest{
		RootPath:        projected.artifactRoot,
		BundleID:        projected.bundleID,
		ResolvedInput:   projected.resolvedInput,
		CandidateReport: candidateReport,
		ScoreEvidence:   scoreEvidence,
		ObjectiveResult: &objectiveResult,
		RenderedReport:  renderedReport,
		CreatedAt:       projected.createdAt,
	})
	if err != nil {
		return Result{}, &Error{Phase: PhaseBundleWriteFailed, Err: err}
	}

	return Result{
		ManifestPath:    manifestPath,
		Bundle:          bundleRef,
		ReportID:        candidateReport.ID,
		CandidateReport: candidateReport,
		ScoreEvidence:   scoreEvidence,
		ObjectiveResult: &objectiveResult,
	}, nil
}

func prepareScorePKL(projected projectedRun, scoreEvidence score.ScoreEvidenceDocument) (string, score.ObjectiveEvidenceRef, func(), error) {
	data, err := artifact.MarshalScoreEvidencePKL(scoreEvidence)
	if err != nil {
		return "", score.ObjectiveEvidenceRef{}, func() {}, err
	}

	dir, err := os.MkdirTemp("", "searchbench-localrun-score-*")
	if err != nil {
		return "", score.ObjectiveEvidenceRef{}, func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	path := filepath.Join(dir, "score.pkl")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		cleanup()
		return "", score.ObjectiveEvidenceRef{}, func() {}, err
	}

	ref := score.ObjectiveEvidenceRef{
		Name:       "current",
		BundlePath: string(projected.expectedBundlePath),
		ScorePath:  filepath.Join(string(projected.expectedBundlePath), "score.pkl"),
		ReportPath: filepath.Join(string(projected.expectedBundlePath), "report.json"),
	}
	return path, ref, cleanup, nil
}

func renderReport(projected projectedRun, candidateReport report.CandidateReport) (*artifact.RenderedReport, error) {
	if !projected.renderReport {
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
