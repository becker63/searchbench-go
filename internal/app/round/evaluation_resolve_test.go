package round

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestResolveExampleManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	out, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts"),
		BundleID:           "round-resolve",
		ReportID:           domain.ReportID("report-round-resolve"),
		Now: func() time.Time {
			return time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("resolveEvaluation() error = %v", err)
	}

	if got, want := out.RoundName, "local-ic-vs-jcodemunch-round-001"; got != want {
		t.Fatalf("RoundName = %q, want %q", got, want)
	}
	if got, want := out.Mode, "evaluation"; got != want {
		t.Fatalf("Mode = %q, want %q", got, want)
	}
	if got, want := out.Policies.Incumbent.ID, domain.SystemID("jcodemunch"); got != want {
		t.Fatalf("Incumbent.ID = %q, want %q", got, want)
	}
	if got, want := out.Policies.Challenger.ID, domain.SystemID("iterative-context"); got != want {
		t.Fatalf("Challenger.ID = %q, want %q", got, want)
	}
	if got, want := out.Scoring.ObjectivePath, filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl"); got != want {
		t.Fatalf("ObjectivePath = %q, want %q", got, want)
	}
	if got, want := out.Output.ResolvedPolicyPaths.Challenger, filepath.ToSlash(filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "policies", "challenger_policy.py")); got != want {
		t.Fatalf("challenger policy path = %q, want %q", got, want)
	}
	if got, want := out.Output.BundleWriterRoot, domain.HostPath(filepath.Join(filepath.Dir(manifestPath), "artifacts")); got == want {
		t.Fatalf("BundleWriterRoot unexpectedly ignored override")
	}
	if got, want := out.Output.ReportFormats, []string{"json", "text"}; !reflectStringsEqual(got, want) {
		t.Fatalf("ReportFormats = %v, want %v", got, want)
	}
	if got, want := out.ReportID, domain.ReportID("report-round-resolve"); got != want {
		t.Fatalf("ReportID = %q, want %q", got, want)
	}
}

func TestResolveOptimizeICManifestRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	manifestPath := filepath.Join(repoRoot(t), "configs", "rounds", "optimize-ic", "round.pkl")
	_, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts", "runs"),
		BundleID:           "optimize-ic-example",
	})
	if err == nil || !strings.Contains(err.Error(), ErrUnsupportedMode.Error()) {
		t.Fatalf("resolveEvaluation() error = %v, want unsupported mode", err)
	}
}

func TestResolveManifestRelativePathsAndParentEvidence(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	parentEvidence := filepath.Join(t.TempDir(), "parent-evidence.pkl")
	if err := os.WriteFile(parentEvidence, []byte("schemaVersion = \"searchbench.round_evidence.v1\"\nreportId = \"parent\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(parentEvidence) error = %v", err)
	}

	out, err := resolveEvaluation(context.Background(), evaluationResolveRequest{
		ManifestPath:       filepath.Join(repoRoot(t), "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl"),
		BundleRootOverride: filepath.Join(t.TempDir(), "bundle-root"),
		BundleID:           "with-parent",
		ParentRef: &score.ObjectiveEvidenceRef{
			Name:         "parent",
			BundlePath:   "fixtures/parent-run",
			EvidencePath: "fixtures/parent-run/evidence.pkl",
			ReportPath:   "fixtures/parent-run/round-report.json",
		},
		ParentEvidencePath: parentEvidence,
	})
	if err != nil {
		t.Fatalf("resolveEvaluation() error = %v", err)
	}

	if got, want := out.Output.BundleCollectionPath, domain.HostPath(filepath.Join(string(out.Output.BundleWriterRoot), "games", "code-localization", "rounds")); got != want {
		t.Fatalf("BundleCollectionPath = %q, want %q", got, want)
	}
	if got, want := out.Output.ExpectedBundlePath, domain.HostPath(filepath.Join(string(out.Output.BundleCollectionPath), "with-parent")); got != want {
		t.Fatalf("ExpectedBundlePath = %q, want %q", got, want)
	}
	if out.Scoring.ParentEvidence == nil {
		t.Fatal("ParentEvidence is nil")
	}
	if got, want := out.Scoring.ParentEvidence.EvidencePath, "fixtures/parent-run/evidence.pkl"; got != want {
		t.Fatalf("ParentEvidence.EvidencePath = %q, want %q", got, want)
	}
	if got, want := out.Scoring.ParentEvidencePath, parentEvidence; got != want {
		t.Fatalf("ParentEvidencePath = %q, want %q", got, want)
	}
}

func TestRoundPackageDoesNotImportGeneratedBindings(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	dir := filepath.Dir(currentFile)
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, dir, func(info os.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parser.ParseDir() error = %v", err)
	}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.Contains(path, "/internal/adapters/config/pkl/generated") {
					t.Fatalf("app/round import %q leaked generated bindings", path)
				}
			}
		}
	}
}

func reflectStringsEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
