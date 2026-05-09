package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
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

	_, err := Resolve(context.Background(), ResolveRequest{
		ManifestPath:       filepath.Join(repoRoot(t), "configs", "experiments", "optimize-ic", "experiment.pkl"),
		BundleRootOverride: filepath.Join(t.TempDir(), "artifacts", "runs"),
		BundleID:           "localrun-unsupported-mode",
	})
	if err == nil || !strings.Contains(err.Error(), ErrUnsupportedMode.Error()) {
		t.Fatalf("Resolve() error = %v, want unsupported mode", err)
	}
}

func TestFakeComparisonProducesCandidateReport(t *testing.T) {
	t.Parallel()

	fixtureDir := createExperimentFixture(t, "")
	plan, err := Resolve(context.Background(), sampleRequest(fixtureDir).Resolve)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	out, executions, err := runComparison(context.Background(), plan, sampleRequest(fixtureDir))
	if err != nil {
		t.Fatalf("runComparison() error = %v", err)
	}
	if out.ID != plan.ReportID {
		t.Fatalf("report ID = %q, want %q", out.ID, plan.ReportID)
	}
	if len(out.Runs.Incumbent) != 1 || len(out.Runs.Challenger) != 1 {
		t.Fatalf("run counts = baseline:%d candidate:%d, want 1/1", len(out.Runs.Incumbent), len(out.Runs.Challenger))
	}
	if out.Decision.Decision != report.DecisionPromote {
		t.Fatalf("decision = %q, want %q", out.Decision.Decision, report.DecisionPromote)
	}
	if len(executions) != 2 {
		t.Fatalf("len(executions) = %d, want 2", len(executions))
	}
	for _, execution := range executions {
		if execution.Result.Executed == nil {
			t.Fatalf("execution result = %#v, want executed evaluator result", execution.Result)
		}
	}
}

func TestFakeComparisonProjectsScoreEvidence(t *testing.T) {
	t.Parallel()

	fixtureDir := createExperimentFixture(t, "")
	plan, err := Resolve(context.Background(), sampleRequest(fixtureDir).Resolve)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	out, _, err := runComparison(context.Background(), plan, sampleRequest(fixtureDir))
	if err != nil {
		t.Fatalf("runComparison() error = %v", err)
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
		Resolve: ResolveRequest{
			ManifestPath:       filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "experiment.pkl"),
			BundleRootOverride: bundleCollection,
			BundleID:           "localrun-success",
			ReportID:           domain.ReportID("report-localrun-success"),
			Now: func() time.Time {
				return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
			},
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
	if len(result.EvaluatorExecutions) != 2 {
		t.Fatalf("len(EvaluatorExecutions) = %d, want 2", len(result.EvaluatorExecutions))
	}

	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "resolved.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "report.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "score.pkl"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "objective.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "metadata.json"))
	assertFileExists(t, filepath.Join(string(result.Bundle.Path), "COMPLETE"))

	var metadata bundlefs.BundleMetadata
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

	var resolved Plan
	decodeJSONFile(t, filepath.Join(string(result.Bundle.Path), "resolved.json"), &resolved)
	if got, want := resolved.Scoring.ObjectivePath, filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl"); got != want {
		t.Fatalf("resolved objective path = %q, want %q", got, want)
	}
	if got := resolved.Output.ResolvedPolicyPaths.Challenger; got == "" {
		t.Fatal("resolved candidate policy path is empty")
	}
	if resolved.Scoring.ParentEvidence != nil {
		t.Fatalf("resolved parent evidence = %#v, want nil for no-parent run", resolved.Scoring.ParentEvidence)
	}
	if !metadataHasPath(metadata, "objective.json") {
		t.Fatalf("metadata files = %#v, want objective.json present", metadata.Files)
	}

	policySourceBytes, err := os.ReadFile(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "policies", "challenger_policy.py"))
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

func TestMaterializeScoreEvidenceCreatesCurrentTempModule(t *testing.T) {
	t.Parallel()

	fixtureDir := createExperimentFixture(t, "")
	plan, err := Resolve(context.Background(), sampleRequest(fixtureDir).Resolve)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	current := sampleScoreEvidence(t, domain.ReportID("report-materialized-current"))

	materialized, err := materializeScoreEvidence(plan, current)
	if err != nil {
		t.Fatalf("materializeScoreEvidence() error = %v", err)
	}
	defer materialized.Cleanup()

	if got, want := materialized.CurrentRef.Name, "current"; got != want {
		t.Fatalf("CurrentRef.Name = %q, want %q", got, want)
	}
	if got, want := materialized.CurrentRef.ScorePath, filepath.Join(string(plan.Output.ExpectedBundlePath), "evidence.pkl"); got != want {
		t.Fatalf("CurrentRef.ScorePath = %q, want %q", got, want)
	}
	content := string(mustReadFile(t, materialized.CurrentScorePath))
	if !strings.Contains(content, `reportId = "report-materialized-current"`) {
		t.Fatalf("materialized current score missing report id:\n%s", content)
	}
}

func TestMaterializeScoreEvidenceAcceptsExplicitParent(t *testing.T) {
	t.Parallel()

	fixtureDir := createExperimentFixture(t, "")
	parentScorePath := writeScoreModuleForTest(t, fixtureDir, "parent-score.pkl", sampleScoreEvidence(t, domain.ReportID("report-parent")))
	request := sampleRequest(fixtureDir)
	request.Resolve.ParentRef = &score.ObjectiveEvidenceRef{
		Name:       "parent",
		BundlePath: filepath.Join(fixtureDir, "artifacts", "runs", "parent-run"),
		ScorePath:  filepath.Join(fixtureDir, "artifacts", "runs", "parent-run", "score.pkl"),
		ReportPath: filepath.Join(fixtureDir, "artifacts", "runs", "parent-run", "report.json"),
	}
	request.Resolve.ParentScorePath = parentScorePath

	plan, err := Resolve(context.Background(), request.Resolve)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	current := sampleScoreEvidence(t, domain.ReportID("report-current"))

	materialized, err := materializeScoreEvidence(plan, current)
	if err != nil {
		t.Fatalf("materializeScoreEvidence() error = %v", err)
	}
	defer materialized.Cleanup()

	if materialized.ParentRef == nil {
		t.Fatal("ParentRef is nil")
	}
	if got, want := materialized.ParentRef.Name, "parent"; got != want {
		t.Fatalf("ParentRef.Name = %q, want %q", got, want)
	}
	if got, want := materialized.ParentScorePath, parentScorePath; got != want {
		t.Fatalf("ParentScorePath = %q, want %q", got, want)
	}
	if got, want := materialized.ParentRef.ScorePath, filepath.Join(fixtureDir, "artifacts", "runs", "parent-run", "score.pkl"); got != want {
		t.Fatalf("ParentRef.ScorePath = %q, want %q", got, want)
	}
}

func TestRunWithParentEvidenceThreadsObjectiveRefs(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	fixtureDir := createExperimentFixture(t, "")
	parentScorePath := writeScoreModuleForTest(t, fixtureDir, "parent-score.pkl", sampleScoreEvidence(t, domain.ReportID("parent-report")))
	parentRef := &score.ObjectiveEvidenceRef{
		Name:       "parent",
		BundlePath: filepath.Join(fixtureDir, "artifacts", "runs", "parent-run"),
		ScorePath:  filepath.Join(fixtureDir, "artifacts", "runs", "parent-run", "score.pkl"),
		ReportPath: filepath.Join(fixtureDir, "artifacts", "runs", "parent-run", "report.json"),
	}

	bundleCollection := filepath.Join(t.TempDir(), "artifacts", "runs")
	result, err := Run(context.Background(), Request{
		Resolve: ResolveRequest{
			ManifestPath:       filepath.Join(fixtureDir, "experiment.pkl"),
			BundleRootOverride: bundleCollection,
			BundleID:           "localrun-parent",
			ParentRef:          parentRef,
			ParentScorePath:    parentScorePath,
			Now: func() time.Time {
				return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if got, want := len(result.ObjectiveResult.EvidenceRefs), 2; got != want {
		t.Fatalf("len(ObjectiveResult.EvidenceRefs) = %d, want %d", got, want)
	}
	if result.ObjectiveResult.EvidenceRefs[0].Name != "current" && result.ObjectiveResult.EvidenceRefs[1].Name != "current" {
		t.Fatalf("ObjectiveResult.EvidenceRefs = %#v, want current ref", result.ObjectiveResult.EvidenceRefs)
	}
	if result.ObjectiveResult.EvidenceRefs[0].Name != "parent" && result.ObjectiveResult.EvidenceRefs[1].Name != "parent" {
		t.Fatalf("ObjectiveResult.EvidenceRefs = %#v, want parent ref", result.ObjectiveResult.EvidenceRefs)
	}

	var resolved Plan
	decodeJSONFile(t, filepath.Join(string(result.Bundle.Path), "resolved.json"), &resolved)
	if resolved.Scoring.ParentEvidence == nil {
		t.Fatal("resolved parent evidence is nil")
	}
	if got, want := resolved.Scoring.ParentEvidence.ScorePath, filepath.Join(fixtureDir, "artifacts", "runs", "parent-run", "score.pkl"); got != want {
		t.Fatalf("resolved parent score path = %q, want %q", got, want)
	}
}

func TestRunObjectiveFailurePreventsCompletedBundle(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	fixtureDir := createExperimentFixture(t, `final = ""`)
	bundleCollection := filepath.Join(t.TempDir(), "artifacts", "runs")
	_, err := Run(context.Background(), Request{
		Resolve: ResolveRequest{
			ManifestPath:       filepath.Join(fixtureDir, "experiment.pkl"),
			BundleRootOverride: bundleCollection,
			BundleID:           "localrun-objective-failure",
			ReportID:           domain.ReportID("report-localrun-objective-failure"),
			Now: func() time.Time {
				return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
			},
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

func TestRunMissingObjectiveFailsBeforeFinalization(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	fixtureDir := createExperimentFixture(t, "")
	if err := os.Remove(filepath.Join(fixtureDir, "scoring", "localization-objective.pkl")); err != nil {
		t.Fatalf("Remove(objective) error = %v", err)
	}

	bundleCollection := filepath.Join(t.TempDir(), "artifacts", "runs")
	_, err := Run(context.Background(), Request{
		Resolve: ResolveRequest{
			ManifestPath:       filepath.Join(fixtureDir, "experiment.pkl"),
			BundleRootOverride: bundleCollection,
			BundleID:           "localrun-missing-objective",
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	localErr := new(Error)
	if !errors.As(err, &localErr) {
		t.Fatalf("error = %T, want *Error", err)
	}
	if localErr.Phase != PhaseResolvePlanFailed {
		t.Fatalf("phase = %q, want %q", localErr.Phase, PhaseResolvePlanFailed)
	}
	finalDir := filepath.Join(bundleCollection, "localrun-missing-objective")
	if _, statErr := os.Stat(filepath.Join(finalDir, "COMPLETE")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("COMPLETE stat error = %v, want not-exist", statErr)
	}
}

func TestRunMissingParentScoreFailsCleanly(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	fixtureDir := createExperimentFixture(t, "")
	bundleCollection := filepath.Join(t.TempDir(), "artifacts", "runs")
	_, err := Run(context.Background(), Request{
		Resolve: ResolveRequest{
			ManifestPath:       filepath.Join(fixtureDir, "experiment.pkl"),
			BundleRootOverride: bundleCollection,
			BundleID:           "localrun-missing-parent-score",
			ParentRef: &score.ObjectiveEvidenceRef{
				Name:       "parent",
				BundlePath: filepath.Join(fixtureDir, "artifacts", "runs", "parent-run"),
				ScorePath:  filepath.Join(fixtureDir, "missing-parent-score.pkl"),
			},
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	localErr := new(Error)
	if !errors.As(err, &localErr) {
		t.Fatalf("error = %T, want *Error", err)
	}
	if localErr.Phase != PhaseScorePKLFailed {
		t.Fatalf("phase = %q, want %q", localErr.Phase, PhaseScorePKLFailed)
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
		filepath.Join(repoRoot, "internal", "pure", "domain"),
		filepath.Join(repoRoot, "internal", "pure", "execution"),
		filepath.Join(repoRoot, "internal", "pure", "score"),
		filepath.Join(repoRoot, "internal", "pure", "report"),
		filepath.Join(repoRoot, "internal", "pure", "codegraph"),
		filepath.Join(repoRoot, "internal", "pure", "usage"),
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
		"internal/ports/pipeline",
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
		Resolve: ResolveRequest{
			ManifestPath:       filepath.Join(tempDir, "experiment.pkl"),
			BundleRootOverride: filepath.Join(tempDir, "artifacts", "runs"),
			BundleID:           "localrun-test",
			ReportID:           domain.ReportID("report-localrun-test"),
			Now: func() time.Time {
				return time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
			},
		},
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
	manifestContent := strings.ReplaceAll(string(manifestBytes), `amends "../../schema/SearchBenchRound.pkl"`, `amends "schema/SearchBenchRound.pkl"`)
	if err := os.WriteFile(filepath.Join(root, "experiment.pkl"), []byte(manifestContent), 0o644); err != nil {
		t.Fatalf("WriteFile(experiment) error = %v", err)
	}

	policyBytes, err := os.ReadFile(filepath.Join(repoRoot(t), "configs", "experiments", "local-ic-vs-jcodemunch", "policies", "challenger_policy.py"))
	if err != nil {
		t.Fatalf("ReadFile(policy) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "policies", "challenger_policy.py"), policyBytes, 0o644); err != nil {
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

	for _, name := range []string{"SearchBenchRound.pkl", "SearchBenchObjective.pkl", "SearchBenchObjectiveHelpers.pkl"} {
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

func metadataHasPath(metadata bundlefs.BundleMetadata, path string) bool {
	for _, file := range metadata.Files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func sampleScoreEvidence(t *testing.T, reportID domain.ReportID) score.ScoreEvidenceDocument {
	t.Helper()
	goldHop, err := score.NewMetricEvidence(score.MetricGoldHop, 6, 4)
	if err != nil {
		t.Fatalf("NewMetricEvidence() error = %v", err)
	}
	return score.ScoreEvidenceDocument{
		SchemaVersion: score.EvidenceSchemaVersion,
		ReportID:      reportID,
		LocalizationDistance: score.LocalizationDistanceEvidence{
			GoldHop: &goldHop,
		},
		Usage: score.UsageEvidence{
			Available:    true,
			MeasuredRuns: 1,
			TotalTokens:  42,
		},
		Regressions: score.RegressionEvidenceSummary{
			SevereCount: 0,
		},
		InvalidPredictions: score.InvalidPredictionEvidence{
			Known: true,
			Count: 0,
		},
	}
}

func writeScoreModuleForTest(t *testing.T, root string, name string, evidence score.ScoreEvidenceDocument) string {
	t.Helper()
	data, err := bundlefs.MarshalScoreEvidencePKL(evidence)
	if err != nil {
		t.Fatalf("MarshalScoreEvidencePKL() error = %v", err)
	}
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
	return path
}
