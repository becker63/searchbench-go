package searchbench

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	"github.com/becker63/searchbench-go/internal/app/round"
	"github.com/becker63/searchbench-go/internal/pure/score"
	"github.com/becker63/searchbench-go/internal/testing/reporoot"
)

const (
	envSearchBenchJCodeMunchCommand       = "SEARCHBENCH_JCODEMUNCH_COMMAND"
	envSearchBenchIterativeContextCommand = "SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND"
)

// TestSearchBenchRoundRunCLIE2E exercises bundle/objective/decision wiring against the
// example manifest that declares real JC/IC backends. Evaluator executions may fail when
// MCP launcher env vars are unset; see configs/rounds/fake-local-e2e for an offline fake-backend smoke manifest.
func TestSearchBenchRoundRunCLIE2E(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	monorepo := reporoot.MonorepoRoot(t)
	modRoot := reporoot.GoModuleRoot(t)
	bundleArtifacts := filepath.Join(t.TempDir(), "artifacts")
	manifestPath := filepath.Join(monorepo, "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	binary := filepath.Join(t.TempDir(), "searchbench")

	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/searchbench")
	build.Dir = modRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	ctx := context.Background()
	runCli := exec.CommandContext(ctx, binary, "run", "--manifest="+manifestPath, "--bundle-root="+bundleArtifacts)
	runCli.Dir = modRoot
	runCli.Env = os.Environ()
	out, err := runCli.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI run failed: %v\n%s", err, out)
	}

	var bundlePrefix string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if after, ok := strings.CutPrefix(line, "bundle="); ok {
			bundlePrefix = strings.TrimSpace(after)
			break
		}
	}
	if bundlePrefix == "" {
		t.Fatalf("CLI output missing bundle= line:\n%s", out)
	}

	assertBundleArtifacts(t, bundlePrefix)

	var objective score.ObjectiveResult
	decodeJSON(t, filepath.Join(bundlePrefix, "objective.json"), &objective)
	if objective.Final == "" {
		t.Fatalf("objective.final is empty")
	}
	for _, v := range objective.Values {
		if strings.TrimSpace(string(v.Kind)) == "" {
			t.Fatalf("objective value %q has empty kind", v.Name)
		}
	}

	var resolvedCli round.Plan
	decodeJSON(t, filepath.Join(bundlePrefix, "resolved-round.json"), &resolvedCli)
	if resolvedCli.RoundName == "" {
		t.Fatalf("resolved-round.json RoundName empty: %#v", resolvedCli)
	}

	var metadata bundlefs.BundleMetadata
	decodeJSON(t, filepath.Join(bundlePrefix, "metadata.json"), &metadata)

	pathsSeen := map[string]struct{}{}
	for _, f := range metadata.Files {
		pathsSeen[f.Path] = struct{}{}
	}
	for _, want := range []string{"resolved-round.json", "round-report.json", "evidence.pkl", "objective.json", "decision.json", "COMPLETE"} {
		if _, ok := pathsSeen[want]; !ok {
			t.Fatalf("metadata.json missing artifact path %q; got %+v", want, metadata.Files)
		}
	}
}

// TestSearchBenchRoundRunEngineE2E mirrors TestSearchBenchRoundRunCLIE2E through round.Run;
// it proves resolver + bundle writers for the JC/IC-declared example manifest, not successful MCP-backed runs.
func TestSearchBenchRoundRunEngineE2E(t *testing.T) {
	t.Parallel()

	requirePkl(t)
	root := reporoot.MonorepoRoot(t)
	bundleArtifacts := filepath.Join(t.TempDir(), "artifacts")
	manifestPath := filepath.Join(root, "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")

	rec, err := round.Run(context.Background(), round.Input{
		EvaluationManifestPath: manifestPath,
		BundleRootOverride:     bundleArtifacts,
		RoundID:                "e2e-root-engine",
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("round.Run error = %v", err)
	}
	if rec.RoundResult == nil {
		t.Fatal("missing RoundResult")
	}
	bp := string(rec.RoundResult.Bundle.Path)
	assertBundleArtifacts(t, bp)

	var rd round.Plan
	decodeJSON(t, filepath.Join(bp, "resolved-round.json"), &rd)
	if rd.RoundName == "" {
		t.Fatalf("resolved-round.json RoundName empty: %#v", rd)
	}
	if rec.RoundResult.ObjectiveResult == nil || rec.RoundResult.ObjectiveResult.Final == "" {
		t.Fatalf("objective result missing final: %#v", rec.RoundResult.ObjectiveResult)
	}
	for _, v := range rec.RoundResult.ObjectiveResult.Values {
		if strings.TrimSpace(string(v.Kind)) == "" {
			t.Fatalf("objective value %q has empty kind", v.Name)
		}
	}
}

func TestSearchBenchFakeLocalRoundRunCLIE2E(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	monorepo := reporoot.MonorepoRoot(t)
	modRoot := reporoot.GoModuleRoot(t)
	bundleArtifacts := filepath.Join(t.TempDir(), "artifacts")
	manifestPath := filepath.Join(monorepo, "configs", "rounds", "fake-local-e2e", "round.pkl")
	binary := filepath.Join(t.TempDir(), "searchbench")

	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/searchbench")
	build.Dir = modRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	ctx := context.Background()
	runCli := exec.CommandContext(ctx, binary, "run", "--manifest="+manifestPath, "--bundle-root="+bundleArtifacts)
	runCli.Dir = modRoot
	runCli.Env = envWithoutMCPLauncher(os.Environ())
	out, err := runCli.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI run failed: %v\n%s", err, out)
	}

	var bundlePrefix string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if after, ok := strings.CutPrefix(line, "bundle="); ok {
			bundlePrefix = strings.TrimSpace(after)
			break
		}
	}
	if bundlePrefix == "" {
		t.Fatalf("CLI output missing bundle= line:\n%s", out)
	}

	assertBundleArtifacts(t, bundlePrefix)
	assertRoundReportJSONHasNoFailures(t, filepath.Join(bundlePrefix, "round-report.json"))
}

func TestSearchBenchFakeLocalRoundRunEngineE2E(t *testing.T) {
	requirePkl(t)

	t.Setenv(envSearchBenchJCodeMunchCommand, "")
	t.Setenv(envSearchBenchIterativeContextCommand, "")

	root := reporoot.MonorepoRoot(t)
	bundleArtifacts := filepath.Join(t.TempDir(), "artifacts")
	manifestPath := filepath.Join(root, "configs", "rounds", "fake-local-e2e", "round.pkl")

	rec, err := round.Run(context.Background(), round.Input{
		EvaluationManifestPath: manifestPath,
		BundleRootOverride:     bundleArtifacts,
		RoundID:                "e2e-fake-local-engine",
		Now: func() time.Time {
			return time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("round.Run error = %v", err)
	}
	if rec.RoundResult == nil {
		t.Fatal("missing RoundResult")
	}
	bp := string(rec.RoundResult.Bundle.Path)
	assertBundleArtifacts(t, bp)
	assertAllEvaluatorRunsSucceeded(t, rec.RoundResult.EvaluatorExecutions)
	assertRoundReportJSONHasNoFailures(t, filepath.Join(bp, "round-report.json"))
}

func envWithoutMCPLauncher(environ []string) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		if strings.HasPrefix(kv, envSearchBenchJCodeMunchCommand+"=") ||
			strings.HasPrefix(kv, envSearchBenchIterativeContextCommand+"=") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

func assertAllEvaluatorRunsSucceeded(tb testing.TB, executions []round.EvaluatorExecution) {
	tb.Helper()
	if len(executions) == 0 {
		tb.Fatal("expected at least one evaluator execution")
	}
	for _, ex := range executions {
		if !ex.Result.Success() {
			tb.Fatalf("evaluator run failed: role=%v match=%v run=%v failure=%+v", ex.Role, ex.MatchID, ex.RunID, ex.Result.Failure)
		}
	}
}

func assertRoundReportJSONHasNoFailures(tb testing.TB, path string) {
	tb.Helper()
	var payload struct {
		Failures struct {
			Incumbent  []any `json:"incumbent"`
			Challenger []any `json:"challenger"`
		} `json:"failures"`
	}
	decodeJSON(tb, path, &payload)
	if len(payload.Failures.Incumbent) != 0 || len(payload.Failures.Challenger) != 0 {
		tb.Fatalf("%s has failures: incumbent=%d challenger=%d", path, len(payload.Failures.Incumbent), len(payload.Failures.Challenger))
	}
}

func assertBundleArtifacts(tb testing.TB, bundleDir string) {
	tb.Helper()
	for _, name := range []string{
		"resolved-round.json", "round-report.json", "round-report.txt", "evidence.pkl",
		"objective.json", "decision.json", "metadata.json", "COMPLETE",
	} {
		path := filepath.Join(bundleDir, name)
		if _, err := os.Stat(path); err != nil {
			tb.Fatalf("expected %s to exist: %v", path, err)
		}
	}
}

func decodeJSON(tb testing.TB, path string, v any) {
	tb.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		tb.Fatalf("decode JSON %s: %v", path, err)
	}
}

func requirePkl(tb testing.TB) {
	tb.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		tb.Skip("pkl not installed")
	}
}
