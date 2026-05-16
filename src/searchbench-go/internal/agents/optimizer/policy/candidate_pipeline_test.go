package policy_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	execpipeline "github.com/becker63/searchbench-go/internal/adapters/pipeline/exec"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/localpath"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/materialize"
	"github.com/becker63/searchbench-go/internal/agents/optimizer/policy"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func TestLocalPathProviderPrepareSeed(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "pyproject.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	seed, err := localpath.Provider{Source: src}.PrepareSeed(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if seed.Kind != optimizer.SeedKindLocalPath {
		t.Fatalf("kind: %s", seed.Kind)
	}
	if seed.Identity.Provider != optimizer.SeedProviderLocalPath {
		t.Fatalf("provider: %s", seed.Identity.Provider)
	}
}

func TestValidateProposalInWorkspaceUsesSuppliedRoot(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	writeICLayout(t, src)
	seed := optimizer.WorkspaceSeed{
		ID:   "t",
		Kind: optimizer.SeedKindLocalPath,
		Root: src,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderLocalPath,
			Source:   src,
			Sha256:   "abc",
		},
	}
	mat := materialize.CandidateMaterializer{}
	ws, cleanup, err := mat.Materialize(seed)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	marker := filepath.Join(ws.Root, ".searchbench-supplied-workspace-marker")
	if err := os.WriteFile(marker, []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}

	proposal := optimizer.NextChallengerProposal{
		ArtifactID:   domain.ArtifactID("test-policy"),
		ArtifactName: "candidate_policy.py",
		InterfaceID:  "iface",
		Code:         "def score_fn():\n    return 1\n",
	}
	res, fail := policy.ValidateProposalInWorkspace(context.Background(), ws, proposal)
	if fail != nil && fail.Kind == optimizer.FailureKindPolicyPipelineInfrastructure {
		t.Skip("uv/ic toolchain not available")
	}
	if fail != nil {
		t.Fatalf("unexpected failure: %v", fail)
	}
	if len(res.Results) == 0 || res.Results[0].CWD != ws.Root {
		t.Fatalf("expected stage_policy cwd=%q, got %+v", ws.Root, res.Results)
	}
	if _, err := os.Stat(filepath.Join(ws.Root, proposal.ArtifactName)); err != nil {
		t.Fatalf("staged policy not in supplied workspace: %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("supplied workspace was replaced")
	}
	orig, _ := os.ReadFile(filepath.Join(src, "pyproject.toml"))
	if string(orig) == "" {
		t.Fatal("source must remain unchanged")
	}
}

func TestValidateProposalWithLocalPathSeedDoesNotMutateSource(t *testing.T) {
	t.Parallel()
	icRoot, err := policy.ResolveIterativeContextRoot()
	if err != nil {
		t.Skip(err)
	}
	before, err := os.ReadFile(filepath.Join(icRoot, "pyproject.toml"))
	if err != nil {
		t.Skip(err)
	}
	proposal := optimizer.NextChallengerProposal{
		ArtifactID:   domain.ArtifactID("test-policy"),
		ArtifactName: "candidate_policy.py",
		InterfaceID:  "iface",
		Code:         "def score_fn():\n    return 1\n",
	}
	_, fail := policy.ValidateProposalWithLocalPathSeed(context.Background(), proposal)
	if fail != nil && fail.Kind == optimizer.FailureKindPolicyPipelineInfrastructure {
		t.Skip("uv/ic toolchain not available")
	}
	after, err := os.ReadFile(filepath.Join(icRoot, "pyproject.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("source tree was mutated during validation")
	}
}

func TestAllowlistBlocksUnsafeCommand(t *testing.T) {
	t.Parallel()
	ws := t.TempDir()
	policyFile := filepath.Join(ws, "p.py")
	if err := os.WriteFile(policyFile, []byte("x=1"), 0o644); err != nil {
		t.Fatal(err)
	}
	allow := execpipeline.ICOptimizerAllowlist{
		UvBinary:         "uv",
		WorkspaceRootAbs: ws,
		PolicyFileAbs:    policyFile,
		PolicyID:         "pid",
		Symbol:           "score_fn",
	}
	if allow.Allows([]string{"bash", "-c", "rm -rf /"}) {
		t.Fatal("expected unsafe command blocked")
	}
}

func writeICLayout(t *testing.T, root string) {
	t.Helper()
	for _, p := range []string{
		"pyproject.toml",
		"src/iterative_context/server.py",
		"src/iterative_context/validate_policy.py",
	} {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("# stub\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
}
