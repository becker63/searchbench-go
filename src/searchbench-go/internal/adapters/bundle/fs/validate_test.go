package bundlefs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/report"
)

func TestValidateCompletedBundlePublishedLiveExample(t *testing.T) {
	t.Parallel()
	root := monorepoRoot(t)
	bundle := filepath.Join(root, "configs", "rounds", "live-ic-vs-jcodemunch",
		"artifacts", "games", "code-localization", "rounds", "live-ic-vs-jcodemunch-001")
	if _, err := os.Stat(filepath.Join(bundle, "COMPLETE")); err != nil {
		t.Skip("published live bundle not present")
	}
	canonical, err := ValidateCompletedBundle(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if canonical.Mode != report.ModeValidateBundle {
		t.Fatalf("mode=%s", canonical.Mode)
	}
	if !canonical.Passed {
		t.Fatalf("expected published bundle to pass validation: %+v", canonical)
	}
}

func monorepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "flake.nix")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("monorepo root not found")
		}
		dir = parent
	}
}
