package round

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	configpkl "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/liveconfig"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
	"github.com/becker63/searchbench-go/internal/pure/usage"
)

// LiveModeInput configures repo-owned live Buck modes (#91).
type LiveModeInput struct {
	ManifestPath string
	ArtifactRoot string
	BundlePath   string
	RoundID      string
	RepoRoot     string

	DatasetMaterializeCacheDir string
}

type liveRoundRunner func(context.Context, Input) (Record, error)

var liveRoundRunnerFn liveRoundRunner = Run

// RunLiveSmoke runs one fresh live round and writes canonical reports at the bundle root.
func RunLiveSmoke(ctx context.Context, in LiveModeInput) error {
	prepareLiveEnv(in)
	record, err := runLiveAttempt(ctx, in, in.BundlePath, 1)
	if err != nil {
		return err
	}
	return writeTopLevelCanonical(in, report.ModeLiveSmoke, record, nil, ConsolidationResult{})
}

// RunEvaluateN runs N fresh attempts under attempts/attempt-NNN/ and consolidates (#91).
func RunEvaluateN(ctx context.Context, in LiveModeInput, attempts int) error {
	if attempts < 1 {
		return fmt.Errorf("evaluate_n: attempts must be >= 1")
	}
	prepareLiveEnv(in)
	outcomes, lastRecord, err := runLiveAttempts(ctx, in, attempts)
	if err != nil {
		return err
	}
	consolidated := ConsolidateAttempts(outcomes, ConsolidationInput{
		Requested: attempts,
		Stability: false,
	})
	return writeTopLevelCanonical(in, report.ModeEvaluateN, lastRecord, consolidated.Summary, consolidated)
}

// RunStabilityProbe runs repeated same-input attempts and reports variance only (#91).
func RunStabilityProbe(ctx context.Context, in LiveModeInput, attempts int) error {
	if attempts < 1 {
		return fmt.Errorf("stability_probe: attempts must be >= 1")
	}
	prepareLiveEnv(in)
	outcomes, lastRecord, err := runLiveAttempts(ctx, in, attempts)
	if err != nil {
		return err
	}
	consolidated := ConsolidateAttempts(outcomes, ConsolidationInput{
		Requested: attempts,
		Stability: true,
	})
	return writeTopLevelCanonical(in, report.ModeStabilityProbe, lastRecord, consolidated.Summary, consolidated)
}

func prepareLiveEnv(in LiveModeInput) {
	repoRoot := strings.TrimSpace(in.RepoRoot)
	if repoRoot == "" {
		return
	}
	liveconfig.LoadSecretsOnly(repoRoot)
	liveconfig.LoadDevOverrides(repoRoot)
	liveconfig.ApplyLiveRuntimeDefaults(liveconfig.Default(repoRoot))
}

func runLiveAttempts(ctx context.Context, in LiveModeInput, attempts int) ([]AttemptOutcome, Record, error) {
	outcomes := make([]AttemptOutcome, 0, attempts)
	var last Record
	for i := 1; i <= attempts; i++ {
		attemptRoot := filepath.Join(in.BundlePath, "attempts", fmt.Sprintf("attempt-%03d", i))
		record, err := runLiveAttempt(ctx, in, attemptRoot, i)
		outcome := outcomeFromRecord(i, record, err)
		outcomes = append(outcomes, outcome)
		if err == nil {
			last = record
			if err := writeAttemptCanonical(in, attemptRoot, record, i); err != nil {
				return outcomes, last, err
			}
		}
	}
	return outcomes, last, nil
}

func runLiveAttempt(ctx context.Context, in LiveModeInput, bundleRoot string, attempt int) (Record, error) {
	if err := os.MkdirAll(bundleRoot, 0o755); err != nil {
		return Record{}, err
	}
	roundID := strings.TrimSpace(in.RoundID)
	if roundID == "" {
		roundID = liveconfig.RoundID
	}
	return liveRoundRunnerFn(ctx, Input{
		EvaluationManifestPath:     in.ManifestPath,
		BundleRootOverride:         bundleRoot,
		RoundID:                    roundID,
		DisableRenderReport:        true,
		DatasetMaterializeCacheDir: in.DatasetMaterializeCacheDir,
	})
}

func outcomeFromRecord(attemptID int, record Record, runErr error) AttemptOutcome {
	outcome := AttemptOutcome{
		AttemptID: attemptID,
		Completed: runErr == nil && record.RoundResult != nil,
		Failures:  make(map[string]int),
	}
	if runErr != nil {
		outcome.Failures[string(execution.FailureCategoryInfrastructure)]++
		return outcome
	}
	result := record.RoundResult
	if result == nil {
		outcome.Failures[string(execution.FailureCategoryArtifact)]++
		return outcome
	}

	outcome.Failures = failureCounts(result.RoundReport.Failures)
	outcome.Passed = totalFailureCount(outcome.Failures) == 0 &&
		len(result.RoundReport.Failures.Incumbent) == 0 &&
		len(result.RoundReport.Failures.Challenger) == 0
	outcome.Decision = string(result.RoundReport.Decision.Decision)
	outcome.TotalTokens, outcome.ToolCalls = usageFromReport(result.RoundReport)
	outcome.PredictionFingerprint = challengerPredictionFingerprint(result.RoundReport)
	if result.ObjectiveResult != nil {
		if final, ok := result.ObjectiveResult.FinalValue(); ok {
			outcome.FinalVal = final.Value
		}
	}
	return outcome
}

func writeAttemptCanonical(in LiveModeInput, attemptRoot string, record Record, attempt int) error {
	if record.RoundResult == nil {
		return fmt.Errorf("attempt %d: missing round result", attempt)
	}
	result := record.RoundResult
	extra := liveTelemetryExtra(in, result, attempt)
	return writeCanonicalAtBundle(attemptRoot, report.ModeRoundRun, report.FreshnessFreshLiveRun, result, extra, nil, ConsolidationResult{
		Passed:   len(result.RoundReport.Failures.Incumbent) == 0 && len(result.RoundReport.Failures.Challenger) == 0,
		Decision: string(result.RoundReport.Decision.Decision),
	})
}

func writeTopLevelCanonical(
	in LiveModeInput,
	mode report.Mode,
	record Record,
	summary *report.AttemptSummary,
	consolidated ConsolidationResult,
) error {
	if err := os.MkdirAll(in.BundlePath, 0o755); err != nil {
		return err
	}
	var result *Result
	if record.RoundResult != nil {
		result = record.RoundResult
	}
	extra := liveTelemetryExtra(in, result, 0)
	extra.Attempts = summary
	applyConsolidatedVerdicts(&extra, mode, consolidated)
	if consolidated.Decision != "" {
		extra.Decision = consolidated.Decision
	}
	passed := consolidated.Passed
	decision := consolidated.Decision
	if result != nil && mode == report.ModeLiveSmoke {
		passed = len(result.RoundReport.Failures.Incumbent) == 0 && len(result.RoundReport.Failures.Challenger) == 0
		decision = string(result.RoundReport.Decision.Decision)
	}
	if err := writeCanonicalAtBundle(in.BundlePath, mode, report.FreshnessFreshLiveRun, result, extra, summary, ConsolidationResult{
		Passed:   passed,
		Decision: decision,
	}); err != nil {
		return err
	}
	meta := map[string]any{
		"schema_version": "searchbench.live_bundle.v1",
		"mode":           mode,
		"round_id":       in.RoundID,
		"bundle_path":    in.BundlePath,
	}
	if summary != nil {
		meta["attempts"] = summary
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(in.BundlePath, "metadata.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(in.BundlePath, "COMPLETE"), []byte("complete\n"), 0o644)
}

func writeCanonicalAtBundle(
	bundlePath string,
	mode report.Mode,
	freshness report.Freshness,
	result *Result,
	extra report.CanonicalReport,
	summary *report.AttemptSummary,
	consolidated ConsolidationResult,
) error {
	if extra.Attempts == nil {
		extra.Attempts = summary
	}
	passed := consolidated.Passed
	decision := consolidated.Decision
	var roundReport report.RoundReport
	if result != nil {
		roundReport = result.RoundReport
		if decision == "" {
			decision = string(result.RoundReport.Decision.Decision)
		}
	}
	if decision == "" && mode == report.ModeStabilityProbe {
		decision = "NO_DECISION"
	}
	if decision != "" {
		roundReport.Decision = report.Decision{Decision: report.DecisionKind(decision)}
	}

	files, err := bundlefs.CanonicalArtifacts(
		mode,
		freshness,
		passed,
		liveconfig.RoundID,
		bundlePath,
		roundReport,
		objectiveFromResult(result),
		extra,
	)
	if err != nil {
		return err
	}
	for _, f := range files {
		path := filepath.Join(bundlePath, f.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, f.Content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func objectiveFromResult(result *Result) *score.ObjectiveResult {
	if result == nil {
		return nil
	}
	return result.ObjectiveResult
}

// ValidateLiveManifest validates a Pkl manifest without running a round.
func ValidateLiveManifest(ctx context.Context, manifestPath string) error {
	spec, err := configpkl.LoadFromPath(ctx, manifestPath)
	if err != nil {
		return err
	}
	return configpkl.Validate(spec)
}

func failureCounts(failures domain.Pair[[]execution.RunFailure]) map[string]int {
	counts := make(map[string]int)
	for _, f := range failures.Incumbent {
		counts[string(f.Category)]++
	}
	for _, f := range failures.Challenger {
		counts[string(f.Category)]++
	}
	return counts
}

func totalFailureCount(counts map[string]int) int {
	n := 0
	for _, v := range counts {
		n += v
	}
	return n
}

func usageFromReport(roundReport report.RoundReport) (totalTokens, toolCalls int) {
	for _, run := range roundReport.Runs.Challenger {
		totalTokens += int(run.Execution.Usage.TotalTokens)
	}
	for _, run := range roundReport.Runs.Incumbent {
		totalTokens += int(run.Execution.Usage.TotalTokens)
	}
	return totalTokens, toolCalls
}

func liveTelemetryExtra(in LiveModeInput, result *Result, attempt int) report.CanonicalReport {
	plan := Plan{
		ManifestPath: in.ManifestPath,
		Round:        RoundConfig{ID: in.RoundID},
	}
	var executions []EvaluatorExecution
	var registry *usage.HashRegistry
	if result != nil {
		executions = result.EvaluatorExecutions
		registry = result.HashRegistry
	}
	extra := buildCanonicalTelemetry(plan, executions, registry)
	if attempt > 0 {
		extra.ModelSeed = liveconfig.ModelSeed(in.RoundID, "match-001", "challenger", attempt)
	}
	return extra
}

func applyConsolidatedVerdicts(extra *report.CanonicalReport, mode report.Mode, consolidated ConsolidationResult) {
	if extra == nil || consolidated.Summary == nil {
		return
	}
	switch mode {
	case report.ModeStabilityProbe:
		switch consolidated.Summary.Verdict {
		case string(stabilityStable):
			extra.Verdict = report.VerdictStable
		case string(stabilityUnstable):
			extra.Verdict = report.VerdictUnstable
		default:
			extra.Verdict = report.VerdictFail
		}
	case report.ModeEvaluateN:
		switch consolidated.Summary.Verdict {
		case verdictPromote:
			extra.PromotionVerdict = report.PromotionVerdictPromote
		case verdictPass, verdictReview:
			extra.PromotionVerdict = report.PromotionVerdictHold
		default:
			extra.PromotionVerdict = report.PromotionVerdictInsufficient
		}
	}
}

func challengerPredictionFingerprint(roundReport report.RoundReport) string {
	var files []string
	for _, run := range roundReport.Runs.Challenger {
		for _, f := range run.Execution.Prediction.Files {
			files = append(files, string(f))
		}
	}
	if len(files) == 0 {
		return ""
	}
	sort.Strings(files)
	return strings.Join(files, "|")
}
