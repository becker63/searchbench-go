package policy_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/adapters/workspace/buckdescriptor"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/materialize"
	"github.com/becker63/searchbench-go/internal/agents/optimizer/policy"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func TestBuckProviderMapsDescriptorToSeed(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	icSrc := filepath.Join(repo, "src", "iterative-context")
	writeICLayout(t, icSrc)
	desc := filepath.Join(icSrc, "optimizable_backend.json")
	if err := os.WriteFile(desc, []byte(`{
  "kind": "searchbench.optimizable_backend.v1",
  "source": {"kind": "local_path", "path": "src/iterative-context", "declared_by": "//src/iterative-context:optimizable_backend"},
  "launcher": {"kind": "mcp_stdio", "cwd_mode": "candidate_workspace", "argv": ["uv","run","python","-m","iterative_context.server"], "env": {}},
  "candidate_validator": {"kind": "ic_policy_pipeline", "steps": ["stage_policy"]},
  "runtime_admin": {"install_tool": "install_score", "verify_tool": "verify_score", "hidden_from_evaluator": true}
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	p := buckdescriptor.Provider{DescriptorPath: desc, RepoRoot: repo}
	seed, err := p.PrepareSeed(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if seed.Kind != optimizer.SeedKindBuckDescriptor {
		t.Fatalf("kind: %s", seed.Kind)
	}
	if seed.Identity.Provider != optimizer.SeedProviderBuckDescriptor {
		t.Fatalf("provider: %s", seed.Identity.Provider)
	}
}

func TestValidateProposalInWorkspaceUsesSuppliedRoot(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	writeICLayout(t, src)
	seed := optimizer.WorkspaceSeed{
		ID:   "t",
		Kind: optimizer.SeedKindBuckDescriptor,
		Root: src,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderBuckDescriptor,
			Source:   "//src/iterative-context:optimizable_backend",
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
		t.Fatalf("expected stage_policy cwd=%q", ws.Root)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("supplied workspace was replaced")
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
