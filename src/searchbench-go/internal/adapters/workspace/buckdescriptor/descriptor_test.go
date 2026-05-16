package buckdescriptor_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/adapters/workspace/buckdescriptor"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

const sampleDescriptor = `{
  "kind": "searchbench.optimizable_backend.v1",
  "source": {
    "kind": "local_path",
    "path": "ic-src",
    "declared_by": "//src/iterative-context:optimizable_backend"
  },
  "launcher": {
    "kind": "mcp_stdio",
    "cwd_mode": "candidate_workspace",
    "argv": ["uv", "run", "python", "-m", "iterative_context.server"],
    "env": {}
  },
  "candidate_validator": {
    "kind": "ic_policy_pipeline",
    "steps": ["stage_policy", "pytest"]
  },
  "runtime_admin": {
    "install_tool": "install_score",
    "verify_tool": "verify_score",
    "hidden_from_evaluator": true
  }
}`

func TestParseDescriptorRejectsRepoChecks(t *testing.T) {
	t.Parallel()
	bad := sampleDescriptor[:len(sampleDescriptor)-1] + `,"repo_checks":{"fast":"//src/iterative-context:check"}}`
	_, err := buckdescriptor.ParseDescriptorJSON([]byte(bad))
	if err == nil {
		t.Fatal("expected repo_checks rejection")
	}
}

func TestParseDescriptorOK(t *testing.T) {
	t.Parallel()
	desc, err := buckdescriptor.ParseDescriptorJSON([]byte(sampleDescriptor))
	if err != nil {
		t.Fatal(err)
	}
	if desc.Source.Kind != "local_path" {
		t.Fatalf("source kind: %s", desc.Source.Kind)
	}
}

func TestBuckProviderMapsDescriptorToSeed(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	icSrc := filepath.Join(repo, "ic-src")
	if err := os.MkdirAll(filepath.Join(icSrc, "src", "iterative_context"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(icSrc, "pyproject.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	descPath := filepath.Join(repo, "descriptor.json")
	if err := os.WriteFile(descPath, []byte(sampleDescriptor), 0o644); err != nil {
		t.Fatal(err)
	}
	p := buckdescriptor.Provider{DescriptorPath: descPath, RepoRoot: repo}
	seed, err := p.PrepareSeed(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if seed.Kind != optimizer.SeedKindBuckDescriptor {
		t.Fatalf("kind: %s", seed.Kind)
	}
	if seed.Archive != "" {
		t.Fatalf("archive must be empty, got %q", seed.Archive)
	}
	if seed.Identity.Provider != optimizer.SeedProviderBuckDescriptor {
		t.Fatalf("provider: %s", seed.Identity.Provider)
	}
}

func TestLoadCommittedDescriptor(t *testing.T) {
	t.Parallel()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Walk up to repo root from package dir.
	repo := wd
	for i := 0; i < 8; i++ {
		candidate := filepath.Join(repo, "src", "iterative-context", "optimizable_backend.json")
		if _, statErr := os.Stat(candidate); statErr == nil {
			desc, err := buckdescriptor.LoadDescriptorFile(candidate)
			if err != nil {
				t.Fatal(err)
			}
			if desc.RepoChecks != nil {
				t.Fatal("committed descriptor must not include repo_checks")
			}
			return
		}
		repo = filepath.Dir(repo)
	}
	t.Skip("optimizable_backend.json not found from test cwd")
}
