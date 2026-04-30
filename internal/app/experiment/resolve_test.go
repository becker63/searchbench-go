package experiment

import (
	"context"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
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

	manifestPath := filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "experiment.pkl")
	out, err := Resolve(context.Background(), Request{
		ManifestPath:       manifestPath,
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts", "runs"),
		BundleID:           "experiment-resolve",
		ReportID:           domain.ReportID("report-experiment-resolve"),
		Now: func() time.Time {
			return time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if got, want := out.ExperimentName, "local-ic-vs-jcodemunch-lca-dev"; got != want {
		t.Fatalf("ExperimentName = %q, want %q", got, want)
	}
	if got, want := out.Mode, "evaluator_only"; got != want {
		t.Fatalf("Mode = %q, want %q", got, want)
	}
	if got, want := out.Scoring.ObjectivePath, filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl"); got != want {
		t.Fatalf("ObjectivePath = %q, want %q", got, want)
	}
	if got, want := out.Output.ResolvedPolicyPaths.Candidate, filepath.ToSlash(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "policies", "candidate_policy.py")); got != want {
		t.Fatalf("candidate policy path = %q, want %q", got, want)
	}
	if got, want := out.Systems.Candidate.Runtime.MaxSteps, 8; got != want {
		t.Fatalf("candidate MaxSteps = %d, want %d", got, want)
	}
	if got, want := out.Systems.Candidate.Runtime.MaxToolCalls, 24; got != want {
		t.Fatalf("candidate MaxToolCalls = %d, want %d", got, want)
	}
	if got, want := out.ReportID, domain.ReportID("report-experiment-resolve"); got != want {
		t.Fatalf("ReportID = %q, want %q", got, want)
	}
}

func TestResolveManifestRelativePathsAndParentEvidence(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	parentScore := filepath.Join(t.TempDir(), "parent-score.pkl")
	if err := os.WriteFile(parentScore, []byte("schemaVersion = \"searchbench.score_evidence.v1\"\nreportId = \"parent\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(parentScore) error = %v", err)
	}

	out, err := Resolve(context.Background(), Request{
		ManifestPath:       filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "experiment.pkl"),
		BundleRootOverride: filepath.Join(t.TempDir(), "bundle-root"),
		BundleID:           "with-parent",
		ParentRef: &score.ObjectiveEvidenceRef{
			Name:       "parent",
			BundlePath: "fixtures/parent-run",
			ScorePath:  "fixtures/parent-run/score.pkl",
			ReportPath: "fixtures/parent-run/report.json",
		},
		ParentScorePath: parentScore,
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if got, want := out.Output.BundleCollectionPath, domain.HostPath(filepath.Join(string(out.Output.BundleWriterRoot), "runs")); got != want {
		t.Fatalf("BundleCollectionPath = %q, want %q", got, want)
	}
	if got, want := out.Output.ExpectedBundlePath, domain.HostPath(filepath.Join(string(out.Output.BundleCollectionPath), "with-parent")); got != want {
		t.Fatalf("ExpectedBundlePath = %q, want %q", got, want)
	}
	if out.Scoring.ParentEvidence == nil {
		t.Fatal("ParentEvidence is nil")
	}
	if got, want := out.Scoring.ParentEvidence.ScorePath, "fixtures/parent-run/score.pkl"; got != want {
		t.Fatalf("ParentEvidence.ScorePath = %q, want %q", got, want)
	}
	if got, want := out.Scoring.ParentScorePath, parentScore; got != want {
		t.Fatalf("ParentScorePath = %q, want %q", got, want)
	}
}

func TestResolvedExperimentPackageDoesNotImportGeneratedBindings(t *testing.T) {
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
					t.Fatalf("app/experiment import %q leaked generated bindings", path)
				}
			}
		}
	}
}

func requirePkl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs(repo root) error = %v", err)
	}
	return root
}
