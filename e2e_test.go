package searchbench

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	"github.com/becker63/searchbench-go/internal/app/round"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestSearchBenchRoundRunCLIE2E(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	repoRoot := repoRootDir(t)
	bundleArtifacts := filepath.Join(t.TempDir(), "artifacts")
	manifestPath := filepath.Join(repoRoot, "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	binary := filepath.Join(t.TempDir(), "searchbench")

	build := exec.Command("go", "build", "-trimpath", "-o", binary, "./cmd/searchbench")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	ctx := context.Background()
	runCli := exec.CommandContext(ctx, binary, "run", "--manifest="+manifestPath, "--bundle-root="+bundleArtifacts)
	runCli.Dir = repoRoot
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

func TestSearchBenchRoundRunEngineE2E(t *testing.T) {
	t.Parallel()

	requirePkl(t)
	root := repoRootDir(t)
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
}

func assertBundleArtifacts(tb testing.TB, bundleDir string) {
	tb.Helper()
	for _, name := range []string{
		"resolved-round.json", "round-report.json", "evidence.pkl",
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

func repoRootDir(tb testing.TB) string {
	tb.Helper()
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		tb.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}
