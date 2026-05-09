package bundlefs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestRoundWritesDurableBundle(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	ref, err := WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	if got, want := ref.BundleID, request.BundleID; got != want {
		t.Fatalf("BundleID = %q, want %q", got, want)
	}
	assertFileExists(t, filepath.Join(string(ref.Path), "resolved-round.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "round-report.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "evidence.pkl"))
	assertFileExists(t, filepath.Join(string(ref.Path), "decision.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "metadata.json"))
	assertFileExists(t, filepath.Join(string(ref.Path), "round-report.md"))
	assertFileExists(t, filepath.Join(string(ref.Path), completeMarkerName))
}

func TestWriteBundleSerializesProvidedRoundEvidence(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	request.RoundEvidence.ChallengerUsage.TotalTokens = 999
	ref, err := WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	evidencePkl := string(mustReadFile(t, filepath.Join(string(ref.Path), "evidence.pkl")))
	if !strings.Contains(evidencePkl, "totalTokens = 999") {
		t.Fatalf("evidence.pkl = %q, want provided usage totalTokens value", evidencePkl)
	}
	if !strings.Contains(evidencePkl, "localizationDistance {") {
		t.Fatalf("evidence.pkl = %q, want Pkl-native camelCase evidence fields", evidencePkl)
	}
}

func TestWriteBundleSerializesObjectiveWhenProvided(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequestWithObjective(t)
	ref, err := WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	assertFileExists(t, filepath.Join(string(ref.Path), "objective.json"))

	var got score.ObjectiveResult
	decodeJSONFile(t, filepath.Join(string(ref.Path), "objective.json"), &got)
	if got.ObjectiveID != request.ObjectiveResult.ObjectiveID {
		t.Fatalf("objective.json objective_id = %q, want %q", got.ObjectiveID, request.ObjectiveResult.ObjectiveID)
	}
	if len(got.Values) != len(request.ObjectiveResult.Values) {
		t.Fatalf("len(objective.json values) = %d, want %d", len(got.Values), len(request.ObjectiveResult.Values))
	}
	if findObjectiveValue(t, got.Values, "regressionPenalty").Value != 1.0 {
		t.Fatalf("objective.json missing preserved named penalty value: %#v", got.Values)
	}
	if final, ok := got.FinalValue(); !ok || final.Name != "final" || final.Value != 0.77 {
		t.Fatalf("objective.json final value = %#v, want final=0.77", final)
	}
}

func TestWriteBundleOmitsObjectiveWhenAbsent(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	ref, err := WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(string(ref.Path), "objective.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("objective.json stat error = %v, want not-exist", err)
	}
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

func TestWriteBundleObjectiveSerializationIsDeterministic(t *testing.T) {
	t.Parallel()

	requestOne := sampleBundleRequestWithObjective(t)
	requestTwo := sampleBundleRequestWithObjective(t)
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

	left := mustReadFile(t, filepath.Join(string(refOne.Path), "objective.json"))
	right := mustReadFile(t, filepath.Join(string(refTwo.Path), "objective.json"))
	if !bytes.Equal(left, right) {
		t.Fatal("objective.json differs between deterministic bundle writes")
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

	files := []string{"resolved-round.json", "round-report.json", "round-report.md", "evidence.pkl", "decision.json", "metadata.json", completeMarkerName}
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
	finalDir := filepath.Join(string(request.RootPath), "games", "code-localization", "rounds", request.BundleID)
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

func TestWriteBundleRejectsMissingRoundEvidence(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	request.RoundEvidence = score.RoundEvidenceDocument{}

	_, err := WriteBundle(context.Background(), request)
	if err == nil {
		t.Fatal("expected error")
	}
	bundleErr := new(Error)
	if !errors.As(err, &bundleErr) {
		t.Fatalf("error = %T, want *Error", err)
	}
	if got, want := bundleErr.Kind, FailureKindValidationFailed; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}
}

func TestSerializationFailureDoesNotCreateCompletedBundle(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	w := newWriter()
	w.marshalScorePKL = func(score.RoundEvidenceDocument) ([]byte, error) {
		return nil, errors.New("fixture score serialization failed")
	}

	_, err := w.WriteBundle(context.Background(), request)
	if err == nil {
		t.Fatal("expected error")
	}
	finalDir := filepath.Join(string(request.RootPath), "games", "code-localization", "rounds", request.BundleID)
	if _, statErr := os.Stat(filepath.Join(finalDir, completeMarkerName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("COMPLETE stat error = %v, want not-exist", statErr)
	}
}

func TestInvalidObjectiveResultFailsBeforeFinalization(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequestWithObjective(t)
	request.ObjectiveResult.Final = "missing"

	_, err := WriteBundle(context.Background(), request)
	if err == nil {
		t.Fatal("expected error")
	}
	bundleErr := new(Error)
	if !errors.As(err, &bundleErr) {
		t.Fatalf("error = %T, want *Error", err)
	}
	if got, want := bundleErr.Kind, FailureKindValidationFailed; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}
	finalDir := filepath.Join(string(request.RootPath), "games", "code-localization", "rounds", request.BundleID)
	if _, statErr := os.Stat(filepath.Join(finalDir, completeMarkerName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("COMPLETE stat error = %v, want not-exist", statErr)
	}
}

func TestObjectiveSerializationFailureDoesNotCreateCompletedBundle(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequestWithObjective(t)
	w := newWriter()
	w.marshalJSON = func(v any) ([]byte, error) {
		if _, ok := v.(score.ObjectiveResult); ok {
			return nil, errors.New("fixture objective serialization failed")
		}
		return marshalDeterministic(v)
	}

	_, err := w.WriteBundle(context.Background(), request)
	if err == nil {
		t.Fatal("expected error")
	}
	finalDir := filepath.Join(string(request.RootPath), "games", "code-localization", "rounds", request.BundleID)
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
	wantPaths := []string{"resolved-round.json", "round-report.json", "round-report.md", "evidence.pkl", "decision.json", "metadata.json", completeMarkerName}
	slices.Sort(gotPaths)
	slices.Sort(wantPaths)
	if !slices.Equal(gotPaths, wantPaths) {
		t.Fatalf("metadata files = %v, want %v", gotPaths, wantPaths)
	}
}

func TestMetadataIncludesObjectiveOnlyWhenPresent(t *testing.T) {
	t.Parallel()

	withObjective, err := WriteBundle(context.Background(), sampleBundleRequestWithObjective(t))
	if err != nil {
		t.Fatalf("WriteBundle(with objective) error = %v", err)
	}
	var withMetadata BundleMetadata
	decodeJSONFile(t, filepath.Join(string(withObjective.Path), "metadata.json"), &withMetadata)
	if !metadataHasPath(withMetadata, "objective.json") {
		t.Fatalf("metadata files = %#v, want objective.json present", withMetadata.Files)
	}

	withoutObjective, err := WriteBundle(context.Background(), sampleBundleRequest(t))
	if err != nil {
		t.Fatalf("WriteBundle(without objective) error = %v", err)
	}
	var withoutMetadata BundleMetadata
	decodeJSONFile(t, filepath.Join(string(withoutObjective.Path), "metadata.json"), &withoutMetadata)
	if metadataHasPath(withoutMetadata, "objective.json") {
		t.Fatalf("metadata files = %#v, want objective.json absent", withoutMetadata.Files)
	}
}

func TestReportSafeOutputsDoNotLeakPolicySource(t *testing.T) {
	t.Parallel()

	request := sampleBundleRequest(t)
	ref, err := WriteBundle(context.Background(), request)
	if err != nil {
		t.Fatalf("WriteBundle() error = %v", err)
	}

	rawSource := "def score(task):\n    return 'candidate'\n"
	for _, name := range []string{"round-report.json", "evidence.pkl", "metadata.json", "round-report.md", "objective.json", "decision.json"} {
		path := filepath.Join(string(ref.Path), name)
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}
		content := string(mustReadFile(t, path))
		if strings.Contains(content, rawSource) {
			t.Fatalf("%s leaked raw policy source", name)
		}
	}
	if strings.Contains(string(mustReadFile(t, filepath.Join(string(ref.Path), "round-report.json"))), `"source"`) {
		t.Fatal(`round-report.json leaked policy source field`)
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
		"tracing",
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

func TestArtifactPackageNoLongerDefinesScoreEvidenceTypes(t *testing.T) {
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
	}, 0)
	if err != nil {
		t.Fatalf("parser.ParseDir() error = %v", err)
	}

	forbiddenTypeNames := map[string]struct{}{
		"ScoreEvidence":        {},
		"MetricEvidence":       {},
		"RoleCounts":           {},
		"ObjectiveResult":      {},
		"ObjectiveValue":       {},
		"ObjectiveEvidenceRef": {},
		"ObjectiveBounds":      {},
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				for _, spec := range gen.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, forbidden := forbiddenTypeNames[typeSpec.Name.Name]; forbidden {
						t.Fatalf("artifact package still defines forbidden score evidence type %q", typeSpec.Name.Name)
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
	taskOne := sampleTask(domain.MatchID("task-1"), domain.RepoRelPath("pkg/bug1.go"))
	taskTwo := sampleTask(domain.MatchID("task-2"), domain.RepoRelPath("pkg/bug2.go"))
	tasks := domain.NewNonEmpty(taskOne, taskTwo)
	spec := report.NewComparisonSpec(domain.NewPair(baseline, candidate), tasks)

	runs := domain.NewPair(
		[]score.ScoredRun{
			sampleScoredRun(t, domain.RoleIncumbent, baseline, taskOne.ID, 4, 5, 0.40, 0.60, 0.30),
			sampleScoredRun(t, domain.RoleIncumbent, baseline, taskTwo.ID, 3, 4, 0.45, 0.55, 0.35),
		},
		[]score.ScoredRun{
			sampleScoredRun(t, domain.RoleChallenger, candidate, taskOne.ID, 1, 1, 0.90, 0.10, 0.95),
			sampleScoredRun(t, domain.RoleChallenger, candidate, taskTwo.ID, 2, 2, 0.80, 0.20, 0.85),
		},
	)
	failures := domain.NewPair(
		[]run.RunFailure{{RunID: domain.RunID("baseline-failure-1"), MatchID: taskTwo.ID, System: baseline.ID, Stage: run.FailureExecute, Message: "baseline retry exhausted"}},
		[]run.RunFailure{},
	)

	roundReport := report.NewRoundReport(
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
				MatchID:   taskTwo.ID,
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
	roundReport.CreatedAt = time.Date(2026, 4, 26, 15, 4, 5, 0, time.UTC)

	roundEvidence, err := report.BuildRoundEvidence(roundReport)
	if err != nil {
		t.Fatalf("BuildRoundEvidence() error = %v", err)
	}
	roundEvidence.GameID = "code-localization"
	roundEvidence.RoundID = "bundle-2026-04-26-fixed"

	return BundleRequest{
		RootPath: domain.HostPath(filepath.Join(t.TempDir(), "artifacts")),
		BundleID: "bundle-2026-04-26-fixed",
		ResolvedInput: map[string]any{
			"manifest_path": "configs/experiments/example/experiment.pkl",
			"round_name":    "bundle-writer-test",
			"mode":          "evaluation",
			"policies":      domain.NewPair(baseline.Ref(), candidate.Ref()),
			"matches":       tasks,
		},
		RoundReport:   roundReport,
		RoundEvidence: roundEvidence,
		RenderedReport: &RenderedReport{
			FileName: "round-report.md",
			Content:  "# Round Report\n\nReview challenger evidence before advancing.\n",
		},
		CreatedAt: time.Date(2026, 4, 26, 16, 0, 0, 0, time.UTC),
	}
}

func sampleBundleRequestWithObjective(t *testing.T) BundleRequest {
	t.Helper()

	request := sampleBundleRequest(t)
	request.ObjectiveResult = sampleObjectiveResult()
	return request
}

func sampleObjectiveResult() *score.ObjectiveResult {
	min := 0.0
	max := 1.0

	return &score.ObjectiveResult{
		SchemaVersion: score.ObjectiveSchemaVersion,
		ObjectiveID:   "candidate_vs_parent_v1",
		EvidenceRefs: []score.ObjectiveEvidenceRef{
			{
				Name:       "current",
				BundlePath: "artifacts/runs/current",
				ScorePath:  "artifacts/runs/current/score.pkl",
				SHA256:     "abc123",
			},
			{
				Name:       "parent",
				BundlePath: "artifacts/runs/parent",
				ScorePath:  "artifacts/runs/parent/score.pkl",
				ReportPath: "artifacts/runs/parent/report.json",
			},
		},
		Values: []score.ObjectiveValue{
			{Name: "currentLocalizationQuality", Value: 0.82, Kind: score.ObjectiveValueIntermediate},
			{Name: "parentLocalizationQuality", Value: 0.74, Kind: score.ObjectiveValueIntermediate},
			{Name: "improvementVsParent", Value: 0.08, Kind: score.ObjectiveValueIntermediate},
			{Name: "tokenEfficiency", Value: 0.91, Kind: score.ObjectiveValueIntermediate},
			{Name: "base", Value: 0.77, Kind: score.ObjectiveValueIntermediate},
			{Name: "regressionPenalty", Value: 1.0, Kind: score.ObjectiveValuePenalty},
			{Name: "invalidPredictionPenalty", Value: 1.0, Kind: score.ObjectiveValuePenalty},
			{Name: "final", Value: 0.77, Kind: score.ObjectiveValueFinal},
		},
		Final:  "final",
		Bounds: &score.ObjectiveBounds{Min: &min, Max: &max},
	}
}

func metadataHasPath(metadata BundleMetadata, path string) bool {
	for _, file := range metadata.Files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func findObjectiveValue(t *testing.T, values []score.ObjectiveValue, name string) score.ObjectiveValue {
	t.Helper()

	for _, value := range values {
		if value.Name == name {
			return value
		}
	}
	t.Fatalf("objective value %q not found", name)
	return score.ObjectiveValue{}
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

func sampleTask(id domain.MatchID, gold domain.RepoRelPath) domain.MatchSpec {
	return domain.MatchSpec{
		ID:        id,
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("repo/example"),
		},
		Input: domain.MatchInput{
			Title: "Fix regression",
			Body:  "The candidate should identify the buggy file.",
		},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{gold},
		},
	}
}

func sampleScoredRun(
	t *testing.T,
	role domain.Role,
	system domain.SystemSpec,
	taskID domain.MatchID,
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
