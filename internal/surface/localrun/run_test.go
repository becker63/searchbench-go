package localrun

import (
	"context"
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/artifact"
	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/surface/config"
)

func TestLocalManifestCanBeLoadedAndValidated(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	manifestPath := filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "experiment.pkl")
	experiment, err := config.ResolveFromPath(context.Background(), manifestPath)
	if err != nil {
		t.Fatalf("ResolveFromPath() error = %v", err)
	}
	if err := config.Validate(experiment); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestRunRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	writeLocalFixtureFiles(t, temp)
	experiment := sampleExperiment(temp)
	experiment.Mode = config.ModeWriterOptimization

	_, err := projectFakeRun(filepath.Join(temp, "experiment.pkl"), experiment, sampleRequest(temp))
	if err == nil || !strings.Contains(err.Error(), errUnsupportedMode.Error()) {
		t.Fatalf("projectFakeRun() error = %v, want unsupported mode", err)
	}
}

func TestFakeComparisonProducesCandidateReport(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	writeLocalFixtureFiles(t, temp)
	projected, err := projectFakeRun(filepath.Join(temp, "experiment.pkl"), sampleExperiment(temp), sampleRequest(temp))
	if err != nil {
		t.Fatalf("projectFakeRun() error = %v", err)
	}

	out, err := runFakeComparison(context.Background(), projected)
	if err != nil {
		t.Fatalf("runFakeComparison() error = %v", err)
	}
	if out.ID != projected.reportID {
		t.Fatalf("report ID = %q, want %q", out.ID, projected.reportID)
	}
	if len(out.Runs.Baseline) != 1 || len(out.Runs.Candidate) != 1 {
		t.Fatalf("run counts = baseline:%d candidate:%d, want 1/1", len(out.Runs.Baseline), len(out.Runs.Candidate))
	}
	if out.Decision.Decision != report.DecisionPromote {
		t.Fatalf("decision = %q, want %q", out.Decision.Decision, report.DecisionPromote)
	}
}

func TestFakeComparisonProjectsScoreEvidence(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	writeLocalFixtureFiles(t, temp)
	projected, err := projectFakeRun(filepath.Join(temp, "experiment.pkl"), sampleExperiment(temp), sampleRequest(temp))
	if err != nil {
		t.Fatalf("projectFakeRun() error = %v", err)
	}
	out, err := runFakeComparison(context.Background(), projected)
	if err != nil {
		t.Fatalf("runFakeComparison() error = %v", err)
	}

	evidence, err := report.ProjectScoreEvidence(out)
	if err != nil {
		t.Fatalf("ProjectScoreEvidence() error = %v", err)
	}
	if err := evidence.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if evidence.LocalizationDistance.GoldHop == nil {
		t.Fatalf("LocalizationDistance = %#v, want gold hop evidence", evidence.LocalizationDistance)
	}
	if evidence.Usage.TotalTokens == 0 {
		t.Fatalf("Usage = %#v, want measured total tokens", evidence.Usage)
	}
}

func TestRunSuccessfulLocalExperimentWritesBundle(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	bundleCollection := filepath.Join(t.TempDir(), "artifacts", "runs")
	result, err := Run(context.Background(), Request{
		ManifestPath:       filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "experiment.pkl"),
		BundleRootOverride: bundleCollection,
		BundleID:           "localrun-success",
		ReportID:           domain.ReportID("report-localrun-success"),
		Now: func() time.Time {
			return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Bundle.Path == "" {
		t.Fatal("bundle path is empty")
	}
	if result.ReportID != domain.ReportID("report-localrun-success") {
		t.Fatalf("ReportID = %q", result.ReportID)
	}
	if result.ObjectiveResult.ObjectiveID == "" {
		t.Fatalf("ObjectiveResult = %#v, want populated objective result", result.ObjectiveResult)
	}

	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "resolved.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "report.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "score.pkl"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "objective.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "metadata.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "COMPLETE"))

	var metadata artifact.BundleMetadata
	decodeJSONFile(t, filepath.Join(string(result.Bundle.Path), "metadata.json"), &metadata)
	gotPaths := make([]string, 0, len(metadata.Files))
	for _, file := range metadata.Files {
		gotPaths = append(gotPaths, file.Path)
	}
	wantPaths := []string{"resolved.json", "report.json", "score.pkl", "objective.json", "metadata.json", "COMPLETE", "report.txt"}
	slices.Sort(gotPaths)
	slices.Sort(wantPaths)
	if !slices.Equal(gotPaths, wantPaths) {
		t.Fatalf("metadata files = %v, want %v", gotPaths, wantPaths)
	}

	policySourceBytes, err := os.ReadFile(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "policies", "candidate_policy.py"))
	if err != nil {
		t.Fatalf("ReadFile(policy) error = %v", err)
	}
	rawPolicySource := string(policySourceBytes)
	for _, name := range []string{"report.json", "score.pkl", "metadata.json", "report.txt", "objective.json"} {
		content := string(mustReadFile(t, filepath.Join(string(result.Bundle.Path), name)))
		if strings.Contains(content, rawPolicySource) {
			t.Fatalf("%s leaked raw policy source", name)
		}
	}
}

func TestRunObjectiveFailurePreventsCompletedBundle(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	fixtureDir := createExperimentFixture(t, `final = ""`)
	bundleCollection := filepath.Join(t.TempDir(), "artifacts", "runs")
	_, err := Run(context.Background(), Request{
		ManifestPath:       filepath.Join(fixtureDir, "experiment.pkl"),
		BundleRootOverride: bundleCollection,
		BundleID:           "localrun-objective-failure",
		ReportID:           domain.ReportID("report-localrun-objective-failure"),
		Now: func() time.Time {
			return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	localErr := new(Error)
	if !errors.As(err, &localErr) {
		t.Fatalf("error = %T, want *Error", err)
	}
	if localErr.Phase != PhaseObjectiveFailed {
		t.Fatalf("phase = %q, want %q", localErr.Phase, PhaseObjectiveFailed)
	}
	finalDir := filepath.Join(bundleCollection, "localrun-objective-failure")
	if _, statErr := os.Stat(filepath.Join(finalDir, "COMPLETE")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("COMPLETE stat error = %v, want not-exist", statErr)
	}
}

func TestPurePackagesStillAvoidPklImports(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	dirs := []string{
		filepath.Join(repoRoot, "internal", "domain"),
		filepath.Join(repoRoot, "internal", "run"),
		filepath.Join(repoRoot, "internal", "score"),
		filepath.Join(repoRoot, "internal", "report"),
		filepath.Join(repoRoot, "internal", "compare"),
	}

	for _, dir := range dirs {
		fs := token.NewFileSet()
		pkgs, err := parser.ParseDir(fs, dir, func(info os.FileInfo) bool {
			name := info.Name()
			return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
		}, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parser.ParseDir(%q) error = %v", dir, err)
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					path := strings.Trim(imp.Path.Value, `"`)
					if strings.Contains(path, "github.com/apple/pkl-go") {
						t.Fatalf("pure package import %q leaked pkl-go", path)
					}
				}
			}
		}
	}
}

func TestLocalRunPackageAvoidsRealRuntimeImports(t *testing.T) {
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

	forbiddenSubstrings := []string{
		"cloudwego/eino",
		"internal/backend",
		"internal/pipeline",
		"openai",
		"anthropic",
		"cerebras",
		"openrouter",
		"mcp",
		"net/http",
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				for _, forbidden := range forbiddenSubstrings {
					if strings.Contains(strings.ToLower(path), forbidden) {
						t.Fatalf("forbidden import %q contains %q", path, forbidden)
					}
				}
			}
		}
	}
}

func sampleRequest(tempDir string) Request {
	return Request{
		ManifestPath:       filepath.Join(tempDir, "experiment.pkl"),
		BundleRootOverride: filepath.Join(tempDir, "artifacts", "runs"),
		BundleID:           "localrun-test",
		ReportID:           domain.ReportID("report-localrun-test"),
		Now: func() time.Time {
			return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
		},
	}
}

func sampleExperiment(tempDir string) config.Experiment {
	return config.Experiment{
		Name: "local-fake-experiment",
		Mode: config.ModeEvaluatorOnly,
		Dataset: config.Dataset{
			Kind:   "lca",
			Name:   "JetBrains-Research/lca-bug-localization",
			Config: "py",
			Split:  "dev",
		},
		Systems: config.Systems{
			Baseline: config.System{
				Id:      "baseline-system",
				Name:    "Baseline",
				Backend: config.BackendJCodeMunch,
				PromptBundle: config.PromptBundle{
					Name: "default",
				},
				Runtime: config.Runtime{MaxSteps: 8, TimeoutSeconds: 300},
			},
			Candidate: config.System{
				Id:      "candidate-system",
				Name:    "Candidate",
				Backend: config.BackendIterativeContext,
				PromptBundle: config.PromptBundle{
					Name: "default",
				},
				Runtime: config.Runtime{MaxSteps: 8, TimeoutSeconds: 300},
				Policy: &config.Policy{
					Id:         "candidate-policy",
					Language:   "python",
					Entrypoint: "score",
					Path:       "policies/candidate_policy.py",
				},
			},
		},
		Evaluator: config.Evaluator{
			Model: config.Model{
				Provider: config.ProviderFake,
				Name:     "fake-model",
			},
			Bounds: config.AgentBounds{MaxModelTurns: 8, MaxToolCalls: 24, TimeoutSeconds: 300},
			Retry:  config.RetryPolicy{MaxAttempts: 2, RetryOnModelError: true, RetryOnFinalizationFailure: true, RetryOnInvalidPrediction: true},
		},
		Scoring: config.Scoring{
			Objective: "scoring/localization-objective.pkl",
		},
		OutputConfig: config.Output{
			ReportFormat: config.ReportFormatPretty,
			BundleRoot:   filepath.Join(tempDir, "artifacts", "runs"),
		},
	}
}

func writeLocalFixtureFiles(t *testing.T, root string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(root, "policies"), 0o755); err != nil {
		t.Fatalf("MkdirAll(policies) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "scoring"), 0o755); err != nil {
		t.Fatalf("MkdirAll(scoring) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "policies", "candidate_policy.py"), []byte("def score(task):\n    return ['src/search_target.go']\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "scoring", "localization-objective.pkl"), []byte("objectiveId = \"placeholder\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(objective) error = %v", err)
	}
}

func createExperimentFixture(t *testing.T, objectiveMutation string) string {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "policies"), 0o755); err != nil {
		t.Fatalf("MkdirAll(policies) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "scoring"), 0o755); err != nil {
		t.Fatalf("MkdirAll(scoring) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "schema"), 0o755); err != nil {
		t.Fatalf("MkdirAll(schema) error = %v", err)
	}

	manifestBytes, err := os.ReadFile(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "experiment.pkl"))
	if err != nil {
		t.Fatalf("ReadFile(experiment) error = %v", err)
	}
	manifestContent := strings.ReplaceAll(string(manifestBytes), `amends "../../schema/SearchBenchExperiment.pkl"`, `amends "schema/SearchBenchExperiment.pkl"`)
	if err := os.WriteFile(filepath.Join(root, "experiment.pkl"), []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("WriteFile(experiment) error = %v", err)
	}

	policyBytes, err := os.ReadFile(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "policies", "candidate_policy.py"))
	if err != nil {
		t.Fatalf("ReadFile(policy) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "policies", "candidate_policy.py"), policyBytes, 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}

	objectiveBytes, err := os.ReadFile(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl"))
	if err != nil {
		t.Fatalf("ReadFile(objective) error = %v", err)
	}
	content := string(objectiveBytes)
	content = strings.ReplaceAll(content, `amends "../../../schema/SearchBenchObjective.pkl"`, `amends "../schema/SearchBenchObjective.pkl"`)
	content = strings.ReplaceAll(content, `import "../../../schema/SearchBenchObjectiveHelpers.pkl" as helpers`, `import "../schema/SearchBenchObjectiveHelpers.pkl" as helpers`)
	if objectiveMutation != "" {
		content += "\n" + objectiveMutation + "\n"
	}
	if err := os.WriteFile(filepath.Join(root, "scoring", "localization-objective.pkl"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(objective) error = %v", err)
	}

	for _, name := range []string{"SearchBenchExperiment.pkl", "SearchBenchObjective.pkl", "SearchBenchObjectiveHelpers.pkl"} {
		schemaBytes, err := os.ReadFile(filepath.Join(repoRoot(t), "configs", "schema", name))
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(root, "schema", name), schemaBytes, 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}
	return root
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

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	return data
}

func decodeJSONFile(t *testing.T, path string, target any) {
	t.Helper()
	if err := json.Unmarshal(mustReadFile(t, path), target); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", path, err)
	}
}
