package materialize

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func TestMaterializeCopiesAndExcludes(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "pyproject.toml"), "[project]\nname='ic'\n")
	mustWrite(t, filepath.Join(src, "src", "iterative_context", "server.py"), "x=1\n")
	mustMkdir(t, filepath.Join(src, ".git"))
	mustMkdir(t, filepath.Join(src, "__pycache__"))
	mustWrite(t, filepath.Join(src, "repomix-output.xml"), "<xml/>")

	digest, err := DigestTree(src)
	if err != nil {
		t.Fatal(err)
	}
	seed := optimizer.WorkspaceSeed{
		ID:   "seed-test",
		Kind: optimizer.SeedKindLocalPath,
		Root: src,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderLocalPath,
			Source:   src,
			Sha256:   digest,
		},
	}
	mat := CandidateMaterializer{}
	ws, cleanup, err := mat.Materialize(seed)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cleanup() }()

	if _, err := os.Stat(filepath.Join(ws.Root, "pyproject.toml")); err != nil {
		t.Fatalf("expected pyproject in candidate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.Root, ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected .git excluded, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(ws.Root, "repomix-output.xml")); !os.IsNotExist(err) {
		t.Fatalf("expected repomix excluded, err=%v", err)
	}
	if ws.Seed.Sha256 != digest {
		t.Fatalf("seed identity not preserved: %+v", ws.Seed)
	}

	mutant := filepath.Join(ws.Root, "src", "iterative_context", "server.py")
	if err := os.WriteFile(mutant, []byte("mutated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, err := os.ReadFile(filepath.Join(src, "src", "iterative_context", "server.py"))
	if err != nil {
		t.Fatal(err)
	}
	if string(orig) == "mutated\n" {
		t.Fatal("source tree must not be mutated")
	}
}

func TestMaterializeCleanupRemovesWorkspace(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "pyproject.toml"), "x\n")
	seed := optimizer.WorkspaceSeed{
		ID:   "seed-cleanup",
		Kind: optimizer.SeedKindLocalPath,
		Root: src,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderLocalPath,
			Source:   src,
			Sha256:   "abc",
		},
	}
	mat := CandidateMaterializer{}
	ws, cleanup, err := mat.Materialize(seed)
	if err != nil {
		t.Fatal(err)
	}
	root := ws.Root
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("expected workspace removed, err=%v", err)
	}
}

func TestMaterializeUniqueWorkspaceIDsPerMaterialization(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "pyproject.toml"), "x\n")
	seed := optimizer.WorkspaceSeed{
		ID:   "seed-unique",
		Kind: optimizer.SeedKindLocalPath,
		Root: src,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderLocalPath,
			Source:   src,
			Sha256:   "abc",
		},
	}
	mat := CandidateMaterializer{}
	ws1, cleanup1, err := mat.Materialize(seed)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cleanup1() }()
	ws2, cleanup2, err := mat.Materialize(seed)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cleanup2() }()
	if ws1.ID == ws2.ID {
		t.Fatal("candidate workspace IDs must be unique per materialization")
	}
	if ws1.Root == ws2.Root {
		t.Fatal("candidate workspace roots must differ")
	}
	if ws1.Seed != ws2.Seed {
		t.Fatal("seed identity should remain stable across materializations")
	}
}

func TestMaterializePreservesExecutableFileMode(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	script := filepath.Join(src, "buck_pytest.sh")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env bash\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := optimizer.WorkspaceSeed{
		ID:   "seed-mode",
		Kind: optimizer.SeedKindLocalPath,
		Root: src,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderLocalPath,
			Source:   src,
			Sha256:   "abc",
		},
	}
	mat := CandidateMaterializer{}
	ws, cleanup, err := mat.Materialize(seed)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cleanup() }()
	info, err := os.Stat(filepath.Join(ws.Root, "buck_pytest.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("mode = %o, want 0755", info.Mode().Perm())
	}
}

func TestMaterializeKeepPreservesWorkspace(t *testing.T) {
	t.Parallel()
	src := t.TempDir()
	mustWrite(t, filepath.Join(src, "pyproject.toml"), "x\n")
	seed := optimizer.WorkspaceSeed{
		ID:   "seed-keep",
		Kind: optimizer.SeedKindLocalPath,
		Root: src,
		Identity: optimizer.WorkspaceSeedIdentity{
			Provider: optimizer.SeedProviderLocalPath,
			Source:   src,
			Sha256:   "abc",
		},
	}
	mat := CandidateMaterializer{Opts: Options{Keep: true}}
	ws, cleanup, err := mat.Materialize(seed)
	if err != nil {
		t.Fatal(err)
	}
	root := ws.Root
	if err := cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("expected keep to preserve workspace: %v", err)
	}
	_ = os.RemoveAll(root)
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
