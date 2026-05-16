package localpath_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/adapters/workspace/localpath"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func TestLocalPathProviderPrepareSeed(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "pyproject.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	p := localpath.Provider{Source: src}
	seed, err := p.PrepareSeed(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if seed.Kind != optimizer.SeedKindLocalPath {
		t.Fatalf("kind: %s", seed.Kind)
	}
	if seed.Identity.Provider != optimizer.SeedProviderLocalPath {
		t.Fatalf("provider: %s", seed.Identity.Provider)
	}
	if seed.Root != src {
		abs, _ := filepath.Abs(src)
		if seed.Root != abs {
			t.Fatalf("root: %s", seed.Root)
		}
	}
}
