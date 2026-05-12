package round

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
)

func TestRunFromScratchWritesContinuationBundle(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	result, err := runEvaluation(context.Background(), evaluationRequest{
		Resolve: evaluationResolveRequest{
			ManifestPath:       filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl"),
			BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
			BundleID:           "round-001",
			Now: func() time.Time {
				return time.Date(2026, 5, 12, 13, 0, 0, 0, time.UTC)
			},
		},
	})
	if err != nil {
		t.Fatalf("runEvaluation() error = %v", err)
	}

	for _, name := range []string{
		"resolved-round.json",
		"round-report.json",
		"evidence.pkl",
		"decision.json",
		"objective.json",
		"continuation.json",
		"continuation.pkl",
		"metadata.json",
		"COMPLETE",
		filepath.Join("policies", "challenger_policy.py"),
	} {
		if _, err := os.Stat(filepath.Join(string(result.Bundle.Path), name)); err != nil {
			t.Fatalf("os.Stat(%q) error = %v", name, err)
		}
	}

	continuation, err := bundlefs.LoadContinuation(result.Bundle.Path)
	if err != nil {
		t.Fatalf("LoadContinuation() error = %v", err)
	}
	if got, want := continuation.SurvivingCandidate.Role, domainSystemRoleChallenger(); got != want {
		t.Fatalf("SurvivingCandidate.Role = %q, want %q", got, want)
	}

	continuationPKLPath := filepath.Join(string(result.Bundle.Path), "continuation.pkl")
	continuationPKL := string(mustReadBundleFile(t, continuationPKLPath))
	for _, needle := range []string{
		`amends "`,
		`code-localization.pkl"`,
		`import "`,
		`code-localization-helpers.pkl" as game`,
		`round = (game.continueFrom(".")) {`,
	} {
		if !strings.Contains(continuationPKL, needle) {
			t.Fatalf("continuation.pkl missing %q:\n%s", needle, continuationPKL)
		}
	}

	var metadata bundlefs.BundleMetadata
	if err := json.Unmarshal(mustReadBundleFile(t, filepath.Join(string(result.Bundle.Path), "metadata.json")), &metadata); err != nil {
		t.Fatalf("Unmarshal(metadata.json) error = %v", err)
	}
	if !metadataHasPath(metadata, "continuation.pkl") {
		t.Fatalf("metadata files = %#v, want continuation.pkl present", metadata.Files)
	}
}

func mustReadBundleFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return data
}

func metadataHasPath(metadata bundlefs.BundleMetadata, want string) bool {
	for _, file := range metadata.Files {
		if file.Path == want {
			return true
		}
	}
	return false
}
