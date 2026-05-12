package round

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestResolveFromScratchManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	out, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
		BundleID:           "round-001",
		Now: func() time.Time {
			return time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("resolveEvaluation() error = %v", err)
	}

	if got, want := out.RoundName, "example-local-ic-vs-jcodemunch-round-001"; got != want {
		t.Fatalf("RoundName = %q, want %q", got, want)
	}
	if got, want := out.CandidateInterfaceID, "iterative_context.selection_policy.v1"; got != want {
		t.Fatalf("CandidateInterfaceID = %q, want %q", got, want)
	}
	if out.Lineage.Continues != "" {
		t.Fatalf("Lineage.Continues = %q, want empty", out.Lineage.Continues)
	}
	if got, want := out.Output.ResolvedPolicyPaths.Challenger, filepath.ToSlash(filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "policies", "challenger_policy.py")); got != want {
		t.Fatalf("Resolved challenger policy path = %q, want %q", got, want)
	}
	if got, want := out.Matches.Len(), 5; got != want {
		t.Fatalf("matches Len() = %d, want %d", got, want)
	}
}

func TestResolveContinuationManifestInheritsParentContext(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	root := t.TempDir()
	dstDataset := filepath.Join(root, "datasets", "JetBrains-Research_lca-bug-localization", "py", "dev.jsonl")
	if err := os.MkdirAll(filepath.Dir(dstDataset), 0o755); err != nil {
		t.Fatalf("MkdirAll(dataset dir) error = %v", err)
	}
	srcDataset := filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "datasets", "JetBrains-Research_lca-bug-localization", "py", "dev.jsonl")
	srcData, err := os.ReadFile(srcDataset)
	if err != nil {
		t.Fatalf("ReadFile(continuation dataset fixture) error = %v", err)
	}
	if err := os.WriteFile(dstDataset, srcData, 0o644); err != nil {
		t.Fatalf("WriteFile(temp dataset) error = %v", err)
	}

	policiesDir := filepath.Join(root, "policies")
	if err := os.MkdirAll(policiesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(policies) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(policiesDir, "challenger_policy.py"), []byte("def score_fn(node, graph, depth):\n    return 0.0\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}

	manifestPath := filepath.Join(root, "round.pkl")
	manifest := `amends "` + filepath.ToSlash(filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "artifacts", "games", "code-localization", "rounds", "round-001", "continuation.pkl")) + `"
import "` + filepath.ToSlash(filepath.Join(repoRoot(t), "configs", "schema", "games", "code-localization-helpers.pkl")) + `" as game

name = "temp-continuation-round-002"

round {
  id = "round-002"

  challenger = (game.iterativeContext("policies/challenger_policy.py")) {
    selectionPolicy {
      id = "challenger-policy-round-002"
    }
  }
}
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("WriteFile(manifest) error = %v", err)
	}
	out, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
		BundleID:           "round-002",
	})
	if err != nil {
		t.Fatalf("resolveEvaluation() error = %v", err)
	}

	if out.Lineage.Continues == "" {
		t.Fatal("Lineage.Continues is empty")
	}
	if out.Scoring.ParentEvidence == nil {
		t.Fatal("Scoring.ParentEvidence is nil")
	}
	if out.Policies.Incumbent.Policy == nil {
		t.Fatal("continued incumbent policy is nil")
	}
	if out.Policies.Challenger.Policy == nil {
		t.Fatal("continued challenger policy is nil")
	}
	if got, want := out.Policies.Incumbent.Policy.ID, out.Policies.Challenger.Policy.ID; got == want {
		t.Fatalf("continued incumbent policy ID unexpectedly equals challenger policy ID after challenger patch")
	}
	if got, want := out.Policies.Incumbent.ID, domainSystemID("iterative-context"); got != want {
		t.Fatalf("continued incumbent ID = %q, want %q", got, want)
	}
	if got, want := out.Policies.Incumbent.Policy.ID.String(), "challenger-policy-round-001"; got != want {
		t.Fatalf("continued incumbent policy ID = %q, want %q", got, want)
	}
	if got, want := out.Policies.Challenger.Policy.ID.String(), "challenger-policy-round-002"; got != want {
		t.Fatalf("continued challenger policy ID = %q, want %q", got, want)
	}
}

func TestResolveContinuationRejectsMissingCompleteMarker(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatalf("MkdirAll(parent) error = %v", err)
	}

	manifestPath := filepath.Join(root, "round.pkl")
	if err := os.WriteFile(manifestPath, []byte(`amends "`+filepath.ToSlash(filepath.Join(repoRoot(t), "configs", "schema", "games", "code-localization.pkl"))+`"
name = "tmp"
round {
  continues = "parent"
  id = "round-002"
  challenger {
    selectionPolicy {
      id = "challenger-policy-round-002"
      kind = "policy"
      path = "policies/challenger_policy.py"
      implements {
        id = "iterative_context.selection_policy.v1"
      }
    }
  }
}
`), 0o644); err != nil {
		t.Fatalf("WriteFile(manifest) error = %v", err)
	}
	policiesDir := filepath.Join(root, "policies")
	if err := os.MkdirAll(policiesDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(policies) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(policiesDir, "challenger_policy.py"), []byte("def score_fn(node, graph, depth):\n    return 0.0\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}

	_, err := resolveEvaluation(context.Background(), evaluationResolveRequest{ManifestPath: manifestPath})
	if err == nil || !stringsContains(err.Error(), "completed marker is missing") {
		t.Fatalf("resolveEvaluation() error = %v, want missing COMPLETE marker failure", err)
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

func stringsContains(value string, needle string) bool {
	return strings.Contains(value, needle)
}

func domainSystemID(value string) domain.SystemID {
	return domain.SystemID(value)
}
