package round

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestRunEvaluateNWritesConsolidatedReport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "bundle")
	prev := liveRoundRunnerFn
	t.Cleanup(func() { liveRoundRunnerFn = prev })

	liveRoundRunnerFn = func(ctx context.Context, input Input) (Record, error) {
		return fakeLiveRecord(input.BundleRootOverride, true, 0.9, 120, "README.md"), nil
	}

	in := LiveModeInput{
		ManifestPath: filepath.Join(dir, "round.pkl"),
		ArtifactRoot: dir,
		BundlePath:   bundlePath,
		RoundID:      "live-test-001",
		RepoRoot:     dir,
	}
	if err := RunEvaluateN(context.Background(), in, 3); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(bundlePath, "report.json"),
		filepath.Join(bundlePath, "COMPLETE"),
		filepath.Join(bundlePath, "attempts", "attempt-001", "report.json"),
		filepath.Join(bundlePath, "attempts", "attempt-003", "report.json"),
	} {
		assertFile(t, path)
	}
	var canonical report.CanonicalReport
	decodeJSONFile(t, filepath.Join(bundlePath, "report.json"), &canonical)
	if canonical.Mode != report.ModeEvaluateN {
		t.Fatalf("mode=%q", canonical.Mode)
	}
	if canonical.Freshness != report.FreshnessFreshLiveRun {
		t.Fatalf("freshness=%q", canonical.Freshness)
	}
	if canonical.Attempts == nil || canonical.Attempts.Count != 3 {
		t.Fatalf("attempts=%+v", canonical.Attempts)
	}
}

func TestRunStabilityProbeNoDecision(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "bundle")
	prev := liveRoundRunnerFn
	t.Cleanup(func() { liveRoundRunnerFn = prev })

	fp := "README.md"
	liveRoundRunnerFn = func(ctx context.Context, input Input) (Record, error) {
		return fakeLiveRecord(input.BundleRootOverride, true, 1, 100, fp), nil
	}

	in := LiveModeInput{
		ManifestPath: filepath.Join(dir, "round.pkl"),
		BundlePath:   bundlePath,
		RoundID:      "live-test-001",
		RepoRoot:     dir,
	}
	if err := RunStabilityProbe(context.Background(), in, 3); err != nil {
		t.Fatal(err)
	}
	var canonical report.CanonicalReport
	decodeJSONFile(t, filepath.Join(bundlePath, "report.json"), &canonical)
	if canonical.Decision != "NO_DECISION" {
		t.Fatalf("decision=%q", canonical.Decision)
	}
	if canonical.Attempts == nil || canonical.Attempts.Verdict != string(stabilityStable) {
		t.Fatalf("attempts=%+v", canonical.Attempts)
	}
}

func fakeLiveRecord(bundleRoot string, passed bool, finalVal float64, tokens int, predictionFile string) Record {
	pred := domain.Prediction{Files: []domain.RepoRelPath{domain.RepoRelPath(predictionFile)}}
	executed := execution.ExecutedRun{
		Prediction: pred,
		Usage:      domain.UsageSummary{TotalTokens: domain.TokenCount(tokens)},
	}
	sr := score.ScoredRun{Execution: executed}
	roundReport := report.RoundReport{
		ID:   "report-test",
		Runs: domain.NewPair([]score.ScoredRun{sr}, []score.ScoredRun{sr}),
		Decision: report.Decision{
			Decision: report.DecisionPromoteChallenger,
		},
	}
	if !passed {
		roundReport.Failures = domain.NewPair(
			nil,
			[]execution.RunFailure{{Category: execution.FailureCategoryInfrastructure}},
		)
	}
	obj := score.ObjectiveResult{
		SchemaVersion: score.ObjectiveSchemaVersion,
		ObjectiveID:   "localization-v1",
		Final:         string(score.MetricComposite),
		Values: []score.ObjectiveValue{
			{Name: string(score.MetricComposite), Value: finalVal, Kind: score.ObjectiveValueFinal},
		},
	}
	return Record{
		RoundResult: &Result{
			ReportID:        "report-test",
			RoundReport:     roundReport,
			ObjectiveResult: &obj,
		},
		RoundBundle: bundleRoot,
	}
}

func assertFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func decodeJSONFile(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}
