package liveconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInputFingerprintStableForSameInputs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	manifestDir := filepath.Join(dir, "configs", "rounds", RoundName)
	if err := os.MkdirAll(filepath.Join(manifestDir, "datasets", "JetBrains-Research_lca-bug-localization", "py"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(manifestDir, "round.pkl")
	slicePath := DatasetSlicePath(Config{
		ManifestDir: manifestDir,
		LCAConfig:   DefaultLCAConfig,
		LCASplit:    DefaultLCASplit,
	})
	if err := os.WriteFile(manifestPath, []byte("manifest-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(slicePath, []byte("dataset-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		RepoRoot:     dir,
		ManifestDir:  manifestDir,
		ManifestPath: manifestPath,
		LCAConfig:    DefaultLCAConfig,
		LCASplit:     DefaultLCASplit,
	}

	fp1, err := InputFingerprint(cfg)
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := InputFingerprint(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if fp1 == "" || fp1 != fp2 {
		t.Fatalf("fingerprints differ: %q vs %q", fp1, fp2)
	}

	if err := os.WriteFile(slicePath, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	fp3, err := InputFingerprint(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if fp3 == fp1 {
		t.Fatal("expected fingerprint to change when dataset slice changes")
	}
}

func TestModelSeedStable(t *testing.T) {
	t.Parallel()
	a := ModelSeed("round", "match", "challenger", 1)
	b := ModelSeed("round", "match", "challenger", 1)
	c := ModelSeed("round", "match", "challenger", 2)
	if a != b || a == c {
		t.Fatalf("seeds: a=%q b=%q c=%q", a, b, c)
	}
}
