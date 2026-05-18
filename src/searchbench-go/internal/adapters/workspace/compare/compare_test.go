package compare

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/adapters/workspace/buckdescriptor"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/localpath"
	"github.com/becker63/searchbench-go/internal/adapters/workspace/materialize"
	"github.com/becker63/searchbench-go/internal/agents/optimizer/policy"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func TestProviderComparisonSmoke(t *testing.T) {
	icRoot, err := policy.ResolveIterativeContextRoot()
	if err != nil {
		t.Skipf("iterative-context not available: %v", err)
	}
	ctx := context.Background()
	localSeed, err := localpath.Provider{Source: icRoot}.PrepareSeed(ctx)
	if err != nil {
		t.Fatalf("local path seed: %v", err)
	}
	buckSeed, err := buckdescriptor.Provider{RepoRoot: repoRootFromIC(icRoot)}.PrepareSeed(ctx)
	if err != nil {
		t.Fatalf("buck descriptor seed: %v", err)
	}
	mat := materialize.CandidateMaterializer{}
	for name, seed := range map[string]optimizer.WorkspaceSeed{
		"local": localSeed,
		"buck":  buckSeed,
	} {
		ws, cleanup, err := mat.Materialize(seed)
		if err != nil {
			t.Fatalf("%s materialize: %v", name, err)
		}
		defer func() { _ = cleanup() }()
		for _, rel := range []string{
			"pyproject.toml",
			"src/iterative_context/server.py",
			"src/iterative_context/validate_policy.py",
			"tests",
		} {
			if _, err := os.Stat(filepath.Join(ws.Root, rel)); err != nil {
				t.Fatalf("%s missing %s: %v", name, rel, err)
			}
		}
	}
}

func repoRootFromIC(icRoot string) string {
	return filepath.Dir(filepath.Dir(icRoot))
}
