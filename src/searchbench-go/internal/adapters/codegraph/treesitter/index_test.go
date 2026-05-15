//go:build cgo

package treesitter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/adapters/codegraph/treesitter"
	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestBuildDirectory_threeFileChain(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pred.go"), `package p

func PredRoot() {
	Mid()
}
`)
	writeFile(t, filepath.Join(dir, "mid.go"), `package p

func Mid() {
	GoldLeaf()
}
`)
	writeFile(t, filepath.Join(dir, "gold.go"), `package p

func GoldLeaf() {}
`)

	store, err := treesitter.BuildDirectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	g, err := store.Build()
	if err != nil {
		t.Fatal(err)
	}

	pred := domain.CanonicalizePath("pred.go")
	gold := domain.CanonicalizePath("gold.go")

	hops, ok := codegraph.MinCallHopsAcrossFiles(g, []domain.RepoRelPath{pred}, []domain.RepoRelPath{gold})
	if !ok {
		t.Fatal("expected localization hops across pred and gold files")
	}
	if hops != 2 {
		t.Fatalf("expected hop count 2 (PredRoot→Mid→GoldLeaf), got %d", hops)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
