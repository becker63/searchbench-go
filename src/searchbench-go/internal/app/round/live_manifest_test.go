package round

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestLiveICVsJCodeMunchManifestResolves(t *testing.T) {
	requirePkl(t)

	roundDir := filepath.Join(repoRoot(t), "configs", "rounds", "live-ic-vs-jcodemunch")
	jsonlPath := writeStubLCADevJSONL(t, roundDir)
	t.Cleanup(func() { _ = os.Remove(jsonlPath) })
	manifestPath := filepath.Join(roundDir, "round.pkl")
	plan, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
		BundleID:           "live-manifest-resolve",
	})
	if err != nil {
		t.Fatalf("resolveEvaluation: %v", err)
	}

	if plan.Dataset.MaxItems == nil || *plan.Dataset.MaxItems != 1 {
		t.Fatalf("dataset maxItems = %v, want 1", plan.Dataset.MaxItems)
	}
	if got, want := plan.Evaluator.Model.Provider, "cerebras"; got != want {
		t.Fatalf("evaluator provider = %q, want %q", got, want)
	}
	if got, want := plan.Policies.Incumbent.Backend, domain.BackendJCodeMunch; got != want {
		t.Fatalf("incumbent backend = %v, want %v", got, want)
	}
	if got, want := plan.Policies.Challenger.Backend, domain.BackendIterativeContext; got != want {
		t.Fatalf("challenger backend = %v, want %v", got, want)
	}

	policyPath := filepath.Join(roundDir, "policies", "challenger_policy.py")
	if plan.Output.ResolvedPolicyPaths.Challenger != filepath.ToSlash(policyPath) {
		t.Fatalf("challenger policy path = %q, want %q", plan.Output.ResolvedPolicyPaths.Challenger, filepath.ToSlash(policyPath))
	}

	fakeIC := evaluatorfake.LocalEvaluatorDefaultAllowedToolNames()
	for _, name := range plan.Evaluator.ToolPolicy.EffectiveAllowed {
		for _, icOnly := range fakeIC {
			if name == icOnly && len(plan.Evaluator.ToolPolicy.EffectiveAllowed) == len(fakeIC) {
				t.Fatalf("live manifest applied fake IC-only allowlist: %#v", plan.Evaluator.ToolPolicy.EffectiveAllowed)
			}
		}
	}
	if len(plan.Evaluator.ToolPolicy.EffectiveAllowed) != 0 {
		t.Fatalf("expected empty effective allow (backend tool surface), got %#v", plan.Evaluator.ToolPolicy.EffectiveAllowed)
	}

	objectivePath := filepath.Join(roundDir, "scoring", "localization-objective.pkl")
	if _, err := os.Stat(objectivePath); err != nil {
		t.Fatalf("objective module missing: %v", err)
	}
}

func writeStubLCADevJSONL(t *testing.T, manifestDir string) string {
	t.Helper()
	path := filepath.Join(
		manifestDir,
		"datasets",
		"JetBrains-Research_lca-bug-localization",
		"py",
		"dev.jsonl",
	)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := `{"repo_owner":"a","repo_name":"early","base_sha":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","issue_title":"t","issue_body":"b","issue_url":"https://example.test/issues/01","changed_files":["a.go"]}` + "\n"
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
