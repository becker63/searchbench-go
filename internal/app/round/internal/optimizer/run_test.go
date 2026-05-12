package optimizer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"

	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	"github.com/becker63/searchbench-go/internal/testing/modeltest"
)

func TestResolveOptimizationManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	plan, err := Resolve(context.Background(), ResolveRequest{
		ManifestPath: filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl"),
		Now: func() time.Time {
			return time.Date(2026, 5, 8, 20, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if got, want := plan.Target.OutputName, "next_challenger_policy.round-002.py"; got != want {
		t.Fatalf("Target.OutputName = %q, want %q", got, want)
	}
	if got, want := plan.ParentBundle.BundleID, "round-001"; got != want {
		t.Fatalf("ParentBundle.BundleID = %q, want %q", got, want)
	}
	if got, want := filepath.Base(plan.InputPolicy.Path), "challenger_policy.py"; got != want {
		t.Fatalf("InputPolicy.Path = %q, want %q", plan.InputPolicy.Path, want)
	}
	if !strings.Contains(strings.Join(plan.IncludedEvidence, ","), "objective_result") {
		t.Fatalf("IncludedEvidence = %#v, want objective_result", plan.IncludedEvidence)
	}
}

func TestResolveOptimizationManifestAllowsParentBundleOverride(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	override := filepath.Join(t.TempDir(), "parent-bundle")
	if err := os.MkdirAll(override, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}

	plan, err := Resolve(context.Background(), ResolveRequest{
		ManifestPath:             filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl"),
		ParentBundlePathOverride: override,
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got, want := string(plan.ParentBundle.BundlePath), override; got != want {
		t.Fatalf("ParentBundle.BundlePath = %q, want %q", got, want)
	}
}

func TestRunRejectsEvaluationManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	_, err := Run(context.Background(), Request{
		Resolve: ResolveRequest{
			ManifestPath: filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl"),
		},
	})
	if err == nil || !strings.Contains(err.Error(), ErrUnsupportedMode.Error()) {
		t.Fatalf("Run() error = %v, want unsupported mode", err)
	}
}

func TestLoadEvidenceRespectsIncludedKinds(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	plan, err := Resolve(context.Background(), ResolveRequest{
		ManifestPath: filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl"),
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	evidence, err := loadEvidence(plan)
	if err != nil {
		t.Fatalf("loadEvidence() error = %v", err)
	}
	if evidence.ReportSummary == nil || evidence.RoundEvidence == nil || evidence.ObjectiveResult == nil {
		t.Fatalf("evidence = %#v, want report summary, round evidence, objective result", evidence)
	}
	if strings.Contains(evidence.InputPolicy.Source, "oracle") {
		t.Fatalf("input policy source unexpectedly contains oracle data:\n%s", evidence.InputPolicy.Source)
	}
}

func TestRunMissingParentBundleFailsTyped(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	plan, err := Resolve(context.Background(), ResolveRequest{
		ManifestPath: filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl"),
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	plan.ParentBundle.BundlePath = "does/not/exist"

	_, err = loadEvidence(plan)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunSuccessfulOptimizerWritesBundle(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl")
	inputPolicyPath := filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "policies", "challenger_policy.py")
	manifestBefore := mustReadFile(t, manifestPath)
	policyBefore := mustReadFile(t, inputPolicyPath)

	result, err := Run(context.Background(), Request{
		Resolve: ResolveRequest{
			ManifestPath:       manifestPath,
			BundleRootOverride: filepath.Join(t.TempDir(), "artifacts", "optimizer"),
			BundleID:           "optimize-success",
			Now: func() time.Time {
				return time.Date(2026, 5, 8, 20, 15, 0, 0, time.UTC)
			},
		},
		Model: modeltest.NewScriptedModel(modeltest.ScriptedResponse{
			Message: schema.AssistantMessage(`{"artifact_id":"next-challenger-round-002","artifact_name":"next_challenger_policy.round-002.py","interface_id":"iterative_context.selection_policy.v1","code":"def score(task):\n    return []\n","summary":"challenger narrows the search frontier"}`, nil),
		}),
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.BundlePath == "" {
		t.Fatal("bundle path is empty")
	}

	for _, name := range []string{
		"resolved-next-challenger.json",
		"optimizer_prompt.txt",
		"optimizer_result.json",
		"next_challenger_policy.round-002.py",
		"metadata.json",
		"COMPLETE",
	} {
		if _, err := os.Stat(filepath.Join(result.BundlePath, name)); err != nil {
			t.Fatalf("os.Stat(%q) error = %v", name, err)
		}
	}

	if got := string(mustReadFile(t, manifestPath)); got != string(manifestBefore) {
		t.Fatal("manifest was mutated")
	}
	if got := string(mustReadFile(t, inputPolicyPath)); got != string(policyBefore) {
		t.Fatal("input policy was mutated")
	}
}

func TestRunFailureDoesNotWriteComplete(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	result, err := Run(context.Background(), Request{
		Resolve: ResolveRequest{
			ManifestPath:       filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl"),
			BundleRootOverride: filepath.Join(t.TempDir(), "artifacts", "optimizer"),
			BundleID:           "optimize-failure",
		},
		Model: modeltest.NewScriptedModel(modeltest.ScriptedResponse{
			Message: schema.AssistantMessage("{\"artifact_id\":\"next-challenger-round-002\",\"artifact_name\":\"next_challenger_policy.round-002.py\",\"interface_id\":\"iterative_context.selection_policy.v1\",\"code\":\"```python\\ndef score(task):\\n    return []\\n```\"}", nil),
		}),
		RetryPolicy: &pureoptimizer.RetryPolicy{
			MaxAttempts:                  1,
			RetryOnModelError:            true,
			RetryOnToolFailure:           true,
			RetryOnFinalizationFailure:   true,
			RetryOnPolicyPipelineFailure: true,
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var failure *pureoptimizer.Failure
	if !errors.As(err, &failure) {
		t.Fatalf("err = %T, want *optimizer.Failure", err)
	}
	if failure.Kind != pureoptimizer.FailureKindNextChallengerFailed {
		t.Fatalf("failure.Kind = %q, want %q", failure.Kind, pureoptimizer.FailureKindNextChallengerFailed)
	}
	if _, statErr := os.Stat(filepath.Join(result.BundlePath, "COMPLETE")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("COMPLETE stat error = %v, want not exist", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(result.BundlePath, "optimizer_result.json")); statErr != nil {
		t.Fatalf("optimizer_result.json stat error = %v", statErr)
	}
}

func requirePkl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found from test file location")
		}
		dir = parent
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	return data
}

func decodeJSONFixture(t *testing.T, path string, target any) {
	t.Helper()
	if err := json.Unmarshal(mustReadFile(t, path), target); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", path, err)
	}
}
