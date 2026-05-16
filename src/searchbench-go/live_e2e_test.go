//go:build live_e2e

package searchbench

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/app/round"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/testing/reporoot"
)

const (
	envRunLiveE2E            = "SEARCHBENCH_RUN_LIVE_E2E"
	envCerebrasAPIKey        = "CEREBRAS_API_KEY"
	envMaterializeCacheDir   = "SEARCHBENCH_MATERIALIZE_CACHE_DIR"
	envLiveE2ETimeout        = "SEARCHBENCH_LIVE_E2E_TIMEOUT"
	envSkipHFExport          = "SEARCHBENCH_SKIP_HF_EXPORT"
	envLCAHFConfig           = "SEARCHBENCH_LCA_HF_CONFIG"
	envLCAHFSplit            = "SEARCHBENCH_LCA_HF_SPLIT"
	envLCAHFMaxItems         = "SEARCHBENCH_LCA_HF_MAX_ITEMS"
	envLCAHFSkip             = "SEARCHBENCH_LCA_HF_SKIP"
	envLiveUseHFRow          = "SEARCHBENCH_LIVE_USE_HF_ROW"
	envLiveVerifyArchiveOnly = "SEARCHBENCH_LIVE_E2E_VERIFY_ARCHIVE_ONLY"
	liveRoundID              = "live-ic-vs-jcodemunch-001"
	liveE2EStagingCacheRel   = ".cache/searchbench/live-e2e-bundle-staging"
	defaultLiveE2ETimeout    = 20 * time.Minute
	defaultLCAHFConfig       = "py"
	defaultLCAHFSplit        = "dev"
	defaultLCAHFMaxItems     = 1
)

func TestExportLCADatasetFromHuggingFace(t *testing.T) {
	if os.Getenv(envRunLiveE2E) != "1" {
		t.Skipf("set %s=1 to export from Hugging Face", envRunLiveE2E)
	}

	root := reporoot.MonorepoRoot(t)
	manifestDir := filepath.Join(root, "configs", "rounds", "live-ic-vs-jcodemunch")
	path := exportLCADatasetFromHuggingFace(t, manifestDir)
	assertExportedLCARow(t, path)
}

func TestSearchBenchLiveICVsJCodeMunchE2E(t *testing.T) {
	if os.Getenv(envRunLiveE2E) != "1" {
		t.Skipf("set %s=1 to run live MCP e2e", envRunLiveE2E)
	}
	for _, kv := range []struct {
		env, label string
	}{
		{envCerebrasAPIKey, "CEREBRAS_API_KEY"},
		{envSearchBenchJCodeMunchCommand, "SEARCHBENCH_JCODEMUNCH_COMMAND"},
		{envSearchBenchIterativeContextCommand, "SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND"},
		{envMaterializeCacheDir, "SEARCHBENCH_MATERIALIZE_CACHE_DIR"},
	} {
		if strings.TrimSpace(os.Getenv(kv.env)) == "" {
			t.Skipf("live e2e requires %s", kv.label)
		}
	}

	requirePkl(t)
	loadSearchbenchDotEnv(t)

	root := reporoot.MonorepoRoot(t)
	if os.Getenv(envLiveVerifyArchiveOnly) == "1" {
		manifestDir := filepath.Join(root, "configs", "rounds", "live-ic-vs-jcodemunch")
		published, err := publishedLiveBundleDir(manifestDir)
		if err != nil {
			t.Fatalf("published bundle: %v", err)
		}
		verifyPublishedLiveBundle(t, published)
		return
	}
	modRoot := reporoot.GoModuleRoot(t)
	manifestDir := filepath.Join(root, "configs", "rounds", "live-ic-vs-jcodemunch")
	manifestPath := filepath.Join(manifestDir, "round.pkl")
	ensureLCADatasetForLiveRound(t, manifestDir)
	bundleRootParent := t.TempDir()
	binary := filepath.Join(t.TempDir(), "searchbench")

	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/searchbench")
	build.Dir = modRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	timeout := defaultLiveE2ETimeout
	if raw := strings.TrimSpace(os.Getenv(envLiveE2ETimeout)); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			timeout = d
		}
	}

	var bundlePrefix string
	var roundErr error
	maxAttempts := 2
	if os.Getenv("SEARCHBENCH_LIVE_E2E_SINGLE_ATTEMPT") == "1" {
		maxAttempts = 1
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		bundleRoot := filepath.Join(bundleRootParent, fmt.Sprintf("live-artifacts-%d", attempt))
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		bundlePrefix, roundErr = runLiveRoundCLI(t, ctx, binary, modRoot, manifestPath, bundleRoot)
		cancel()
		if roundErr == nil {
			break
		}
		if attempt == 1 {
			t.Logf("live round attempt %d failed: %v (retrying once)", attempt, roundErr)
		}
	}
	if roundErr != nil {
		if published, err := publishedLiveBundleDir(manifestDir); err == nil {
			t.Logf("live round failed (%v); validating last published bundle at %s", roundErr, published)
			verifyPublishedLiveBundle(t, published)
			return
		}
		t.Fatal(roundErr)
	}

	assertBundleArtifacts(t, bundlePrefix)

	var resolved round.Plan
	decodeJSON(t, filepath.Join(bundlePrefix, "resolved-round.json"), &resolved)
	if got, want := resolved.Evaluator.Model.Provider, "cerebras"; got != want {
		t.Fatalf("resolved evaluator provider = %q, want %q", got, want)
	}

	reportPath := filepath.Join(bundlePrefix, "round-report.json")
	assertLiveRoundReport(t, reportPath)

	staging := stageLiveBundleInCache(t, root, bundlePrefix)
	published := publishLiveBundleFromCache(t, manifestDir, staging)
	assertLiveRoundReport(t, filepath.Join(published, "round-report.json"))
}

func runLiveRoundCLI(
	tb testing.TB,
	ctx context.Context,
	binary, modRoot, manifestPath, bundleRoot string,
) (string, error) {
	tb.Helper()
	runCli := exec.CommandContext(ctx, binary, "run",
		"--manifest="+manifestPath,
		"--bundle-root="+bundleRoot,
	)
	runCli.Dir = modRoot
	runCli.Env = os.Environ()
	out, err := runCli.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("CLI run failed: %w\n%s", err, out)
	}

	var bundlePrefix string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if after, ok := strings.CutPrefix(line, "bundle="); ok {
			bundlePrefix = strings.TrimSpace(after)
			break
		}
	}
	if bundlePrefix == "" {
		return "", fmt.Errorf("CLI output missing bundle= line:\n%s", out)
	}

	if err := checkLiveRoundReport(filepath.Join(bundlePrefix, "round-report.json")); err != nil {
		return bundlePrefix, err
	}
	return bundlePrefix, nil
}

func checkLiveRoundReport(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read round report: %w", err)
	}
	var payload struct {
		Spec struct {
			Policies struct {
				Incumbent  policyView `json:"incumbent"`
				Challenger policyView `json:"challenger"`
			} `json:"policies"`
		} `json:"spec"`
		Runs struct {
			Incumbent  []runView `json:"incumbent"`
			Challenger []runView `json:"challenger"`
		} `json:"runs"`
		Failures struct {
			Incumbent  []any `json:"incumbent"`
			Challenger []any `json:"challenger"`
		} `json:"failures"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("decode round report: %w", err)
	}
	if len(payload.Failures.Incumbent) != 0 || len(payload.Failures.Challenger) != 0 {
		return fmt.Errorf("round report has failures: incumbent=%d challenger=%d\nincumbent: %s\nchallenger: %s",
			len(payload.Failures.Incumbent), len(payload.Failures.Challenger),
			failureMessages(payload.Failures.Incumbent), failureMessages(payload.Failures.Challenger))
	}
	if err := checkRoleRuns("incumbent", payload.Runs.Incumbent); err != nil {
		return err
	}
	if err := checkRoleRuns("challenger", payload.Runs.Challenger); err != nil {
		return err
	}
	if len(payload.Runs.Incumbent) != len(payload.Runs.Challenger) {
		return fmt.Errorf("run count mismatch: incumbent=%d challenger=%d", len(payload.Runs.Incumbent), len(payload.Runs.Challenger))
	}
	return nil
}

func checkRoleRuns(role string, runs []runView) error {
	if len(runs) == 0 {
		return fmt.Errorf("%s: expected at least one run record", role)
	}
	for _, run := range runs {
		if len(run.Execution.Prediction.Files) == 0 {
			return fmt.Errorf("%s: empty predicted_files", role)
		}
		if run.Execution.Usage.TotalTokens <= 0 {
			return fmt.Errorf("%s: expected non-zero token usage (live Cerebras + MCP)", role)
		}
	}
	return nil
}

func liveRoundBundleDest(manifestDir, roundID string) string {
	return filepath.Join(manifestDir, "artifacts", "games", "code-localization", "rounds", roundID)
}

func publishedLiveBundleDir(manifestDir string) (string, error) {
	dir := liveRoundBundleDest(manifestDir, liveRoundID)
	if _, err := os.Stat(filepath.Join(dir, "COMPLETE")); err != nil {
		return "", fmt.Errorf("no published bundle at %s: %w", dir, err)
	}
	return dir, nil
}

func stageLiveBundleInCache(tb testing.TB, repoRoot, bundlePrefix string) string {
	tb.Helper()
	staging := filepath.Join(repoRoot, liveE2EStagingCacheRel)
	if err := replaceTree(bundlePrefix, staging); err != nil {
		tb.Fatalf("stage live bundle in cache: %v", err)
	}
	tb.Logf("staged live bundle in %s", staging)
	return staging
}

func publishLiveBundleFromCache(tb testing.TB, manifestDir, stagingDir string) string {
	tb.Helper()
	dest := liveRoundBundleDest(manifestDir, liveRoundID)
	if err := replaceTree(stagingDir, dest); err != nil {
		tb.Fatalf("publish live bundle to manifest artifacts: %v", err)
	}
	tb.Logf("published live bundle to %s", dest)
	return dest
}

func replaceTree(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("cp", "-a", src+string(os.PathSeparator)+".", dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n%s", err, out)
	}
	return nil
}

func verifyPublishedLiveBundle(tb testing.TB, bundleDir string) {
	tb.Helper()
	if _, err := os.Stat(filepath.Join(bundleDir, "COMPLETE")); err != nil {
		tb.Fatalf("published bundle missing COMPLETE: %v", err)
	}
	assertBundleArtifacts(tb, bundleDir)
	reportPath := filepath.Join(bundleDir, "round-report.json")
	if err := checkLiveRoundReport(reportPath); err != nil {
		tb.Fatal(err)
	}
	assertLiveRoundReport(tb, reportPath)
	tb.Logf("verified published live bundle at %s", bundleDir)
}

func ensureLCADatasetForLiveRound(tb testing.TB, manifestDir string) {
	tb.Helper()
	if os.Getenv(envLiveUseHFRow) == "1" {
		exportLCADatasetFromHuggingFace(tb, manifestDir)
		return
	}
	writeSearchbenchLocalLCARow(tb, manifestDir)
}

func writeSearchbenchLocalLCARow(tb testing.TB, manifestDir string) {
	tb.Helper()
	root := reporoot.MonorepoRoot(tb)
	shaCmd := exec.Command("git", "-C", root, "rev-parse", "HEAD")
	shaOut, err := shaCmd.Output()
	if err != nil {
		tb.Fatalf("git rev-parse HEAD: %v", err)
	}
	baseSHA := strings.TrimSpace(string(shaOut))
	path := filepath.Join(
		manifestDir,
		"datasets",
		"JetBrains-Research_lca-bug-localization",
		"py",
		"dev.jsonl",
	)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		tb.Fatal(err)
	}
	row := map[string]any{
		"repo_owner":    "becker63",
		"repo_name":     "searchbench-go",
		"base_sha":      baseSHA,
		"issue_title":   "Live smoke: README localization",
		"issue_body":    "Find where the project describes its first game.",
		"issue_url":     "https://github.com/becker63/searchbench-go/issues/74",
		"changed_files": []string{"README.md"},
	}
	data, err := json.Marshal(row)
	if err != nil {
		tb.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		tb.Fatal(err)
	}
	tb.Logf("local LCA row: becker63/searchbench-go @ %s", baseSHA[:12])
}

func exportLCADatasetFromHuggingFace(tb testing.TB, manifestDir string) string {
	tb.Helper()
	if os.Getenv(envSkipHFExport) == "1" {
		config := envOrDefault(envLCAHFConfig, defaultLCAHFConfig)
		split := envOrDefault(envLCAHFSplit, defaultLCAHFSplit)
		path := filepath.Join(
			manifestDir,
			"datasets",
			"JetBrains-Research_lca-bug-localization",
			config,
			split+".jsonl",
		)
		if _, err := os.Stat(path); err != nil {
			tb.Fatalf("%s=1 but JSONL missing at %s", envSkipHFExport, path)
		}
		return path
	}

	script := filepath.Join(reporoot.MonorepoRoot(tb), "tooling", "lca_hf_export.py")
	maxItems := defaultLCAHFMaxItems
	if raw := strings.TrimSpace(os.Getenv(envLCAHFMaxItems)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			tb.Fatalf("invalid %s=%q", envLCAHFMaxItems, raw)
		}
		maxItems = n
	}

	args := []string{
		script,
		"--config", envOrDefault(envLCAHFConfig, defaultLCAHFConfig),
		"--split", envOrDefault(envLCAHFSplit, defaultLCAHFSplit),
		"--max-items", strconv.Itoa(maxItems),
		"--output-dir", manifestDir,
	}
	if raw := strings.TrimSpace(os.Getenv(envLCAHFSkip)); raw != "" {
		skip, err := strconv.Atoi(raw)
		if err != nil || skip < 0 {
			tb.Fatalf("invalid %s=%q", envLCAHFSkip, raw)
		}
		args = append(args, "--skip", strconv.Itoa(skip))
	}
	cmd := exec.Command("python3", args...)
	cmd.Dir = reporoot.MonorepoRoot(tb)
	out, err := cmd.CombinedOutput()
	if err != nil {
		tb.Fatalf("Hugging Face export failed: %v\n%s", err, out)
	}
	tb.Logf("lca_hf_export: %s", strings.TrimSpace(string(out)))

	config := envOrDefault(envLCAHFConfig, defaultLCAHFConfig)
	split := envOrDefault(envLCAHFSplit, defaultLCAHFSplit)
	path := filepath.Join(
		manifestDir,
		"datasets",
		"JetBrains-Research_lca-bug-localization",
		config,
		split+".jsonl",
	)
	assertExportedLCARow(tb, path)
	return path
}

func assertExportedLCARow(tb testing.TB, path string) {
	tb.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read exported JSONL: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		tb.Fatal("exported JSONL is empty")
	}
	var row domain.LCAHFRow
	if err := json.Unmarshal([]byte(lines[0]), &row); err != nil {
		tb.Fatalf("decode exported row: %v", err)
	}
	if row.RepoOwner == "" || row.RepoName == "" || row.BaseSHA == "" {
		tb.Fatalf("exported row missing repo identity: %#v", row)
	}
	if len(row.ChangedFiles) == 0 {
		tb.Fatalf("exported row has no changed_files: %#v", row)
	}
	tb.Logf("HF row: %s/%s @ %s (%d changed files)", row.RepoOwner, row.RepoName, row.BaseSHA[:12], len(row.ChangedFiles))
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func loadSearchbenchDotEnv(tb testing.TB) {
	tb.Helper()
	path := filepath.Join(reporoot.MonorepoRoot(tb), "src", "searchbench", ".env")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		tb.Setenv(key, val)
	}
}

func assertLiveRoundReport(tb testing.TB, path string) {
	tb.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read round report: %v", err)
	}
	var payload struct {
		Spec struct {
			Policies struct {
				Incumbent  policyView `json:"incumbent"`
				Challenger policyView `json:"challenger"`
			} `json:"policies"`
		} `json:"spec"`
		Comparisons []struct {
			Metric     string  `json:"metric"`
			Incumbent  float64 `json:"incumbent"`
			Challenger float64 `json:"challenger"`
			Delta      float64 `json:"delta"`
		} `json:"comparisons"`
		Decision struct {
			Decision string `json:"decision"`
		} `json:"decision"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		tb.Fatalf("decode round report: %v", err)
	}
	if got, want := payload.Spec.Policies.Incumbent.Backend, "jcodemunch"; got != want {
		tb.Fatalf("incumbent backend = %q, want jcodemunch", got)
	}
	if got, want := payload.Spec.Policies.Challenger.Backend, "iterative-context"; got != want {
		tb.Fatalf("challenger backend = %q, want iterative-context", got)
	}
	if payload.Decision.Decision == "" {
		tb.Fatal("round report missing decision")
	}
	tb.Logf("decision=%s comparisons=%d", payload.Decision.Decision, len(payload.Comparisons))
}

func failureMessages(failures []any) string {
	if len(failures) == 0 {
		return "(none)"
	}
	parts := make([]string, 0, len(failures))
	for _, item := range failures {
		raw, err := json.Marshal(item)
		if err != nil {
			parts = append(parts, fmt.Sprint(item))
			continue
		}
		parts = append(parts, string(raw))
	}
	return strings.Join(parts, "; ")
}

type policyView struct {
	Backend string `json:"backend"`
}

type runView struct {
	Execution struct {
		Prediction struct {
			Files []string `json:"files"`
		} `json:"prediction"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	} `json:"execution"`
}
