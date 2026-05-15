package round

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

func TestFakeE2E_FakeLocalManifestUsesFakeBackends(t *testing.T) {
	requirePkl(t)

	manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "fake-local-e2e", "round.pkl")
	plan, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
		BundleID:           "audit-fake-local-backend-kind",
	})
	if err != nil {
		t.Fatalf("resolveEvaluation: %v", err)
	}
	if got, want := plan.Policies.Incumbent.Backend, domain.BackendFake; got != want {
		t.Fatalf("incumbent backend = %v, want %v", got, want)
	}
	if got, want := plan.Policies.Challenger.Backend, domain.BackendFake; got != want {
		t.Fatalf("challenger backend = %v, want %v", got, want)
	}
}

// TestFakeE2E_LocalManifestIncumbentUsesJCodeMunchBackend documents the example
// manifest's incumbent backend so fake-e2e auditors know JC wiring is declared.
func TestFakeE2E_LocalManifestIncumbentUsesJCodeMunchBackend(t *testing.T) {
	requirePkl(t)

	manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	plan, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
		BundleID:           "audit-backend-kind",
	})
	if err != nil {
		t.Fatalf("resolveEvaluation: %v", err)
	}
	if got, want := plan.Policies.Incumbent.Backend, domain.BackendJCodeMunch; got != want {
		t.Fatalf("incumbent backend = %v, want %v", got, want)
	}
	if got, want := plan.Policies.Challenger.Backend, domain.BackendIterativeContext; got != want {
		t.Fatalf("challenger backend = %v, want %v", got, want)
	}
}

// TestFakeE2E_LocalManifestToolFactoryFailsWithoutMCPEnv proves default prepared
// tools cannot start for JC/IC backends when MCP launcher env vars are empty.
func TestFakeE2E_LocalManifestToolFactoryFailsWithoutMCPEnv(t *testing.T) {
	requirePkl(t)

	t.Setenv(envJCodeMunchCommand, "")
	t.Setenv(envIterativeContextCommand, "")

	manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	plan, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
		BundleID:           "audit-mcp-env",
	})
	if err != nil {
		t.Fatalf("resolveEvaluation: %v", err)
	}
	task := plan.Matches.Head()

	incSpec := run.NewSpec(domain.RunID("audit-inc"), task, plan.Policies.Incumbent)
	ctx := context.Background()
	_, _, err = defaultPreparedToolFactory()(ctx, incSpec)
	if err == nil {
		t.Fatal("expected error preparing jCodeMunch tools without SEARCHBENCH_JCODEMUNCH_COMMAND")
	}

	chSpec := run.NewSpec(domain.RunID("audit-ch"), task, plan.Policies.Challenger)
	_, _, err = defaultPreparedToolFactory()(ctx, chSpec)
	if err == nil {
		t.Fatal("expected error preparing iterative-context tools without SEARCHBENCH_ITERATIVE_CONTEXT_COMMAND")
	}
}
