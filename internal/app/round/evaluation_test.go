package round

import (
	"context"
	"os"
	"path/filepath"
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
}
