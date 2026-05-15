package bundlefs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestLoadContinuationRequiresContinuationJSONEvenWhenPKLExists(t *testing.T) {
	t.Parallel()

	bundleDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(bundleDir, completeMarkerName), []byte(completeMarkerContent), 0o644); err != nil {
		t.Fatalf("WriteFile(COMPLETE) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(bundleDir, continuationPKLFileName), []byte("amends \"../../schema/games/code-localization.pkl\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(continuation.pkl) error = %v", err)
	}

	_, err := LoadContinuation(domain.HostPath(bundleDir))
	if !errors.Is(err, ErrMissingContinuation) {
		t.Fatalf("LoadContinuation() error = %v, want ErrMissingContinuation", err)
	}
}

func TestLoadContinuationRequiresCompleteMarkerEvenWhenPKLExists(t *testing.T) {
	t.Parallel()

	bundleDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(bundleDir, continuationPKLFileName), []byte("amends \"../../schema/games/code-localization.pkl\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(continuation.pkl) error = %v", err)
	}

	_, err := LoadContinuation(domain.HostPath(bundleDir))
	if !errors.Is(err, ErrIncompleteBundle) {
		t.Fatalf("LoadContinuation() error = %v, want ErrIncompleteBundle", err)
	}
}
