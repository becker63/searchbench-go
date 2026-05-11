package round

import (
        "context"
        "os"
        "os/exec"
        "path/filepath"
        "testing"
)

// TestEvaluateMatchesDoesNotWriteCompleteMarker confirms EvaluateMatches stays
// effect-free with respect to the round bundle. The COMPLETE marker is the
// canonical "round bundle is durable" signal; it must only appear after
// WriteBundle runs.
func TestEvaluateMatchesDoesNotWriteCompleteMarker(t *testing.T) {
        t.Parallel()

        resolved, bundleRoot := newTestRound(t)

        if _, err := EvaluateMatches(context.Background(), resolved, Input{
                EvaluationManifestPath: resolved.Round.ManifestPath,
                BundleRootOverride:     bundleRoot,
                DisableRenderReport:    true,
        }); err != nil {
                t.Fatalf("EvaluateMatches: %v", err)
        }

        assertNoCompleteMarker(t, bundleRoot)
}

// TestBuildEvidenceDoesNotWriteCompleteMarker confirms BuildEvidence is a pure
// projection over MatchRecords and produces no bundle artifacts.
func TestBuildEvidenceDoesNotWriteCompleteMarker(t *testing.T) {
        t.Parallel()

        resolved, bundleRoot := newTestRound(t)

        matches, err := EvaluateMatches(context.Background(), resolved, Input{
                EvaluationManifestPath: resolved.Round.ManifestPath,
                BundleRootOverride:     bundleRoot,
                DisableRenderReport:    true,
        })
        if err != nil {
                t.Fatalf("EvaluateMatches: %v", err)
        }

        evidence, err := BuildEvidence(resolved.Game, resolved, matches)
        if err != nil {
                t.Fatalf("BuildEvidence: %v", err)
        }
        if evidence.RoundID == "" {
                t.Fatalf("BuildEvidence produced empty RoundID")
        }
        assertNoCompleteMarker(t, bundleRoot)
}

// TestEvaluateObjectiveDoesNotWriteCompleteMarker confirms EvaluateObjective
// runs the scoring graph without persisting any bundle artifact.
func TestEvaluateObjectiveDoesNotWriteCompleteMarker(t *testing.T) {
        t.Parallel()

        resolved, bundleRoot := newTestRound(t)
        input := Input{
                EvaluationManifestPath: resolved.Round.ManifestPath,
                BundleRootOverride:     bundleRoot,
                DisableRenderReport:    true,
        }

        matches, err := EvaluateMatches(context.Background(), resolved, input)
        if err != nil {
                t.Fatalf("EvaluateMatches: %v", err)
        }
        evidence, err := BuildEvidence(resolved.Game, resolved, matches)
        if err != nil {
                t.Fatalf("BuildEvidence: %v", err)
        }
        if _, err := EvaluateObjective(context.Background(), resolved, evidence, matches, input); err != nil {
                t.Fatalf("EvaluateObjective: %v", err)
        }
        assertNoCompleteMarker(t, bundleRoot)
}

// TestWriteBundleWritesCompleteMarker confirms WriteBundle is the only phase
// that writes the COMPLETE marker for the round bundle.
func TestWriteBundleWritesCompleteMarker(t *testing.T) {
        t.Parallel()

        resolved, bundleRoot := newTestRound(t)
        input := Input{
                EvaluationManifestPath: resolved.Round.ManifestPath,
                BundleRootOverride:     bundleRoot,
                DisableRenderReport:    true,
        }

        matches, err := EvaluateMatches(context.Background(), resolved, input)
        if err != nil {
                t.Fatalf("EvaluateMatches: %v", err)
        }
        evidence, err := BuildEvidence(resolved.Game, resolved, matches)
        if err != nil {
                t.Fatalf("BuildEvidence: %v", err)
        }
        objective, err := EvaluateObjective(context.Background(), resolved, evidence, matches, input)
        if err != nil {
                t.Fatalf("EvaluateObjective: %v", err)
        }

        assertNoCompleteMarker(t, bundleRoot)

        bundleRef, err := WriteBundle(context.Background(), resolved, matches, evidence, objective, input)
        if err != nil {
                t.Fatalf("WriteBundle: %v", err)
        }
        if bundleRef.Path == "" {
                t.Fatalf("WriteBundle returned empty bundle path")
        }
        completeMarker := filepath.Join(string(bundleRef.Path), "COMPLETE")
        if _, err := os.Stat(completeMarker); err != nil {
                t.Fatalf("WriteBundle did not write COMPLETE marker at %s: %v", completeMarker, err)
        }
}

// newTestRound resolves a round plan from the canonical Round001 manifest in
// a fresh temp bundle root. It returns the Resolved record and the override
// bundle root for assertions.
func newTestRound(t *testing.T) (Resolved, string) {
        t.Helper()

        if _, err := exec.LookPath("pkl"); err != nil {
                t.Skip("pkl CLI not available on PATH")
        }

        bundleRoot := t.TempDir()
        manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")

        resolvedGame, err := ResolveGame(context.Background(), Input{})
        if err != nil {
                t.Fatalf("ResolveGame: %v", err)
        }
        resolved, err := ResolveRound(context.Background(), resolvedGame, Input{
                EvaluationManifestPath: manifestPath,
                BundleRootOverride:     bundleRoot,
        })
        if err != nil {
                t.Fatalf("ResolveRound: %v", err)
        }
        return resolved, bundleRoot
}

func assertNoCompleteMarker(t *testing.T, bundleRoot string) {
        t.Helper()

        var found []string
        _ = filepath.Walk(bundleRoot, func(path string, info os.FileInfo, err error) error {
                if err != nil {
                        return nil
                }
                if info.IsDir() {
                        return nil
                }
                if filepath.Base(path) == "COMPLETE" {
                        found = append(found, path)
                }
                return nil
        })
        if len(found) > 0 {
                t.Fatalf("expected no COMPLETE marker, found: %v", found)
        }
}
