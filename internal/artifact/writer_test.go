package artifact

import (
	"bytes"
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

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

func TestWriteBundleSerializesCandidateReport(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	ref, err := WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	if got, want := ref.BundleID, request.BundleID; got != want {
		t.Fatalf("BundleID = %q, want %q", got, want)
	}
	assertFileExists(t, filepath.Join(string(ref.Path), "resolved.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "report.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "score.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "metadata.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "report.md"))
	assertFileExists(t, filepath.Join(string(ref.Path), completeMarkerName))
}

func TestWriteBundleWritesCompleteMarkerLast(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	w := newWriter()
	var writes []string
	w.afterWrite = func(name string) {
		writes = append(writes, name)
	}

	_, err := w.WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	if got, want := writes[len(writes)-1], completeMarkerName; got != want {
		t.Fatalf("last write = %q, want %q (writes=%v)", got, want, writes)
	}
}

func TestWriteBundleDeterministicSerialization(t *testing.T) {
	t.Parallel()

	requestOne := sampleBundleRequest(t)
	requestTwo := sampleBundleRequest(t)
	requestTwo.RootPath = domain.HostPath(filepath.Join(t.TempDir(), "artifacts"))
	requestTwo.BundleID = requestOne.BundleID

	refOne, err := WriteBundle(context.Background(), requestOne)
	if err != nil {
		t.Fatalf("WriteBundle(requestOne) error = %v", err)
	}
	refTwo, err := WriteBundle(context.Background(), requestTwo)
	if err != nil {
		t.Fatalf("WriteBundle(requestTwo) error = %v", err)
	}

	files := []string{"resolved.json", "report.json", "report.md", "score.json", "metadata.json", completeMarkerName}
	for _, name := range files {
		left := mustReadFile(t, filepath.Join(string(refOne.Path), name))
		right := mustReadFile(t, filepath.Join(string(refTwo.Path), name))
		if !bytes.Equal(left, right) {
			t.Fatalf("%s differs between deterministic bundle writes", name)
		}
	}
}

func TestWriteBundleFailsWhenCompletedBundleExists(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	finalDir := filepath.Join(string(request.RootPath), "runs", request.BundleID)
	if err := os.MkdirAll(finalDir, 0o755); err != nil {
		t.Fatalf("os.MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(finalDir, completeMarkerName), []byte(completeMarkerContent), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	_, err := WriteBundle(context.Background(), request)
	if err == nil {
		t.Fatal("expected error")
	}
	bundleErr := new(Error)
	if !errors.As(err, &bundleErr) {
		t.Fatalf("error = %T, want *Error", err)
	}
	if got, want := bundleErr.Kind, FailureKindAlreadyExists; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}
}

func TestSerializationFailureDoesNotCreateCompletedBundle(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	w := newWriter()
	w.marshalJSON = func(v any) ([]byte, error) {
		if _, ok := v.(ScoreEvidence); ok {
			return nil, errors.New("fixture score serialization failed")
		}
		return marshalDeterministic(v)
	}

	_, err := w.WriteBundle(context.Background(), request)
	if err == nil {
		t.Fatal("expected error")
	}
	finalDir := filepath.Join(string(request.RootPath), "runs", request.BundleID)
	if _, statErr := os.Stat(filepath.Join(finalDir, completeMarkerName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("COMPLETE stat error = %v, want not-exist", statErr)
	}
}

func TestMetadataListsEveryGeneratedArtifact(t *testing.T) {
	t.Parallel()

	ref, err := WriteBundle(context.Background(), sampleBundleRequest(t))
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	var metadata BundleMetadata
	decodeJSONFile(t, filepath.Join(string(ref.Path), "metadata.json"), &metadata)

	gotPaths := make([]string, 0, len(metadata.Files))
	for _, file := range metadata.Files {
		gotPaths = append(gotPaths, file.Path)
	}
	wantPaths := []string{"resolved.json", "report.json", "report.md", "score.json", "metadata.json", completeMarkerName}
	slices.Sort(gotPaths)
	slices.Sort(wantPaths)
	if !slices.Equal(gotPaths, wantPaths) {
		t.Fatalf("metadata files = %v, want %v", gotPaths, wantPaths)
	}
}

func TestReportSafeOutputsDoNotLeakPolicySource(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	ref, err := WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	policySource := request.ResolvedInput.Systems.Candidate.Policy
	if policySource == nil {
		t.Fatal("expected candidate policy ref")
	}
	rawSource := "def score(task):\n    return 'candidate'\n"
	for _, name := range []string{"report.json", "score.json", "metadata.json", "report.md"} {
		content := string(mustReadFile(t, filepath.Join(string(ref.Path), name)))
		if strings.Contains(content, rawSource) {
			t.Fatalf("%s leaked raw policy source", name)
		}
	}
	if strings.Contains(string(mustReadFile(t, filepath.Join(string(ref.Path), "report.json"))), `"source"`) {
		t.Fatal(`report.json leaked policy source field`)
	}
}

func TestArtifactPackageAvoidsForbiddenImports(t *testing.T) {
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
		"pkl",
		"cloudwego/eino",
		"mcp",
		"langsmith",
		"langfuse",
		"materialization",
		"tree-sitter",
		"treesitter",
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

func sampleBundleRequest(t *testing.T) BundleRequest {
	t.Helper()

	policySource := "def score(task):\n    return 'candidate'\n"
	baseline := sampleBaselineSystem()
	candidate := sampleCandidateSystem(policySource)
	taskOne := sampleTask(domain.TaskID("task-1"), domain.RepoRelPath("pkg/bug1.go"))
	taskTwo := sampleTask(domain.TaskID("task-2"), domain.RepoRelPath("pkg/bug2.go"))
	tasks := domain.NewNonEmpty(taskOne, taskTwo)
	spec := report.NewComparisonSpec(domain.NewPair(baseline, candidate), tasks)

	runs := domain.NewPair(
		[]score.ScoredRun{
			sampleScoredRun(t, domain.RoleBaseline, baseline, taskOne.ID, 4, 5, 0.40, 0.60, 0.30),
			sampleScoredRun(t, domain.RoleBaseline, baseline, taskTwo.ID, 3, 4, 0.45, 0.55, 0.35),
		},
		[]score.ScoredRun{
			sampleScoredRun(t, domain.RoleCandidate, candidate, taskOne.ID, 1, 1, 0.90, 0.10, 0.95),
			sampleScoredRun(t, domain.RoleCandidate, candidate, taskTwo.ID, 2, 2, 0.80, 0.20, 0.85),
		},
	)
	failures := domain.NewPair(
		[]run.RunFailure{{RunID: domain.RunID("baseline-failure-1"), TaskID: taskTwo.ID, System: baseline.ID, Stage: run.FailureExecute, Message: "baseline retry exhausted"}},
		[]run.RunFailure{},
	)

	candidateReport := report.NewCandidateReport(
		domain.ReportID("report-immutable-bundle"),
		spec,
		runs,
		failures,
		[]report.ScoreComparison{
			report.NewScoreComparison(score.MetricGoldHop, 3.5, 1.5),
			report.NewScoreComparison(score.MetricIssueHop, 4.5, 1.5),
			report.NewScoreComparison(score.MetricTokenEfficiency, 0.425, 0.85),
			report.NewScoreComparison(score.MetricCost, 0.575, 0.15),
			report.NewScoreComparison(score.MetricComposite, 0.325, 0.90),
		},
		[]report.Regression{
			{
				TaskID:    taskTwo.ID,
				Metric:    score.MetricCost,
				Baseline:  0.10,
				Candidate: 0.20,
				Delta:     0.10,
				Severity:  report.RegressionMinor,
				Reason:    "candidate cost is slightly higher on task-2",
			},
		},
		report.PromotionDecision{
			Decision: report.DecisionReview,
			Reason:   "candidate improves core metrics but has a minor cost regression",
		},
	)
	candidateReport.CreatedAt = time.Date(2026, 4, 26, 15, 4, 5, 0, time.UTC)

	return BundleRequest{
		RootPath: domain.HostPath(filepath.Join(t.TempDir(), "artifacts")),
		BundleID: "bundle-2026-04-26-fixed",
		ResolvedInput: ResolvedComparisonInput{
			Systems: domain.NewPair(baseline.Ref(), candidate.Ref()),
			Tasks:   tasks,
			Parallelism: ParallelismConfig{
				Mode:       "sequential",
				MaxWorkers: 1,
			},
			ScoringProfile: "default",
			ReportOptions: ReportOptions{
				Format: "markdown",
			},
		},
		CandidateReport: candidateReport,
		RenderedReport: &RenderedReport{
			FileName: "report.md",
			Content:  "# Candidate Report\n\nPROMOTE? review first.\n",
		},
		CreatedAt: time.Date(2026, 4, 26, 16, 0, 0, 0, time.UTC),
	}
}

func sampleBaselineSystem() domain.SystemSpec {
	return domain.SystemSpec{
		ID:      domain.SystemID("baseline-system"),
		Name:    "Baseline",
		Backend: domain.BackendJCodeMunch,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-baseline",
		},
		PromptBundle: domain.PromptBundleRef{
			Name:    "bundle",
			Version: "v1",
		},
	}
}

func sampleCandidateSystem(policySource string) domain.SystemSpec {
	policy := domain.NewPythonPolicy(domain.PolicyID("policy-1"), policySource, "score")
	return domain.SystemSpec{
		ID:      domain.SystemID("candidate-system"),
		Name:    "Candidate",
		Backend: domain.BackendIterativeContext,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-candidate",
		},
		PromptBundle: domain.PromptBundleRef{
			Name:    "bundle",
			Version: "v2",
		},
		Policy: &policy,
		Runtime: domain.RuntimeConfig{
			MaxSteps:        8,
			MaxOutputTokens: 2048,
		},
	}
}

func sampleTask(id domain.TaskID, gold domain.RepoRelPath) domain.TaskSpec {
	return domain.TaskSpec{
		ID:        id,
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("repo/example"),
		},
		Input: domain.TaskInput{
			Title: "Fix regression",
			Body:  "The candidate should identify the buggy file.",
		},
		Oracle: domain.TaskOracle{
			GoldFiles: []domain.RepoRelPath{gold},
		},
	}
}

func sampleScoredRun(
	t *testing.T,
	role domain.Role,
	system domain.SystemSpec,
	taskID domain.TaskID,
	goldHop score.HopDistance,
	issueHop score.HopDistance,
	efficiency score.EfficiencyScore,
	cost score.CostScore,
	composite score.CompositeScore,
) score.ScoredRun {
	t.Helper()

	task := sampleTask(taskID, domain.RepoRelPath("pkg/bug.go"))
	spec := run.NewSpec(domain.RunID(string(role)+"-"+taskID.String()), task, system)
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("session-"+taskID.String()))
	executed := run.NewExecuted(
		prepared,
		domain.Prediction{Files: []domain.RepoRelPath{"pkg/out.go"}},
		domain.UsageSummary{InputTokens: 10, OutputTokens: 5, TotalTokens: 15},
		domain.TraceID("trace-"+taskID.String()),
		time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 26, 12, 0, 2, 0, time.UTC),
	)

	scores, err := score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: goldHop},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: issueHop},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: efficiency},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: cost},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: composite},
	)
	if err != nil {
		t.Fatalf("NewScoreSet() error = %v", err)
	}

	scored, err := score.NewScoredRun(executed, scores)
	if err != nil {
		t.Fatalf("NewScoredRun() error = %v", err)
	}
	return scored
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
