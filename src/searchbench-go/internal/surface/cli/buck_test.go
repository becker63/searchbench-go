package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/testing/reporoot"
)

func TestBuckRoundValidateMode(t *testing.T) {
	t.Parallel()
	requirePkl(t)

	root := reporoot.MonorepoRoot(t)
	manifest := filepath.Join(root, "configs", "rounds", "live-ic-vs-jcodemunch", "round.pkl")

	var out bytes.Buffer
	err := RunWithWriters(
		context.Background(),
		[]string{
			"--quiet", "buck", "round",
			"--mode=validate",
			"--manifest", manifest,
			"--repo-root", root,
		},
		&out,
		io.Discard,
	)
	if err != nil {
		t.Fatalf("RunWithWriters() error = %v", err)
	}
	if got := out.String(); !bytes.Contains(out.Bytes(), []byte("ok manifest=")) {
		t.Fatalf("output=%q", got)
	}
}

func TestBuckValidateBundleMode(t *testing.T) {
	t.Parallel()
	root := reporoot.MonorepoRoot(t)
	bundle := filepath.Join(root, "configs", "rounds", "live-ic-vs-jcodemunch", "artifacts",
		"games", "code-localization", "rounds", "live-ic-vs-jcodemunch-001")
	if _, err := os.Stat(filepath.Join(bundle, "COMPLETE")); err != nil {
		t.Skip("published live bundle not present")
	}

	var out bytes.Buffer
	err := RunWithWriters(
		context.Background(),
		[]string{
			"--quiet", "buck", "round",
			"--mode=validate_bundle",
			"--bundle-path", bundle,
			"--repo-root", root,
		},
		&out,
		io.Discard,
	)
	if err != nil {
		t.Fatalf("RunWithWriters() error = %v", err)
	}
	if got := out.String(); !bytes.Contains(out.Bytes(), []byte("ok bundle=")) {
		t.Fatalf("output=%q", got)
	}
}

func TestBuckUnknownModeRejected(t *testing.T) {
	t.Parallel()
	err := RunWithWriters(
		context.Background(),
		[]string{"buck", "round", "--mode=not-a-mode", "--repo-root", t.TempDir()},
		io.Discard,
		io.Discard,
	)
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
}
