package logging

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

func TestSystemSpecKVOmitsPolicySource(t *testing.T) {
	t.Parallel()

	system := domain.SystemSpec{
		ID:      domain.SystemID("candidate-system"),
		Name:    "Candidate",
		Backend: domain.BackendIterativeContext,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-candidate",
		},
		PromptBundle: domain.PromptBundleRef{Name: "bundle"},
		Policy:       ptr(domain.NewPythonPolicy(domain.PolicyID("policy-1"), "def score(task): return 'candidate'", "score")),
	}

	kv := SystemSpecKV(system)
	assertNoKey(t, kv, "source")
	assertNoValue(t, kv, "def score(task): return 'candidate'")
	assertHasKey(t, kv, "policy_sha256")
}

func TestTaskKVOmitsOracleFields(t *testing.T) {
	t.Parallel()

	task := domain.TaskSpec{
		ID:        domain.TaskID("task-1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("/tmp/repo"),
		},
		Oracle: domain.TaskOracle{
			GoldFiles: []domain.RepoRelPath{"pkg/bug.go"},
		},
	}

	kv := TaskKV(task)
	assertNoKey(t, kv, "oracle")
	assertNoKey(t, kv, "gold_files")
	assertNoValue(t, kv, domain.RepoRelPath("pkg/bug.go"))
	assertHasKey(t, kv, "task_id")
}

func TestScoreSetKVIncludesRequiredMetrics(t *testing.T) {
	t.Parallel()

	scores, err := score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 1},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 2},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.8},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.2},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.9},
	)
	if err != nil {
		t.Fatalf("NewScoreSet() error = %v", err)
	}

	kv := ScoreSetKV(scores)
	for _, key := range []string{
		"gold_hop",
		"issue_hop",
		"token_efficiency",
		"cost",
		"composite",
	} {
		assertHasKey(t, kv, key)
	}
}

func TestReportSummaryKVIsCompact(t *testing.T) {
	t.Parallel()

	report := report.NewCandidateReport(
		domain.ReportID("report-1"),
		report.ComparisonSpec{},
		domain.NewPair([]score.ScoredRun{{}}, []score.ScoredRun{{}, {}}),
		domain.NewPair([]run.RunFailure{{}}, []run.RunFailure{}),
		[]report.ScoreComparison{{}, {}},
		[]report.Regression{{}},
		report.PromotionDecision{Decision: report.DecisionPromote},
	)

	kv := ReportSummaryKV(report)
	assertHasKey(t, kv, "report_id")
	assertHasKey(t, kv, "baseline_runs")
	assertHasKey(t, kv, "candidate_runs")
	assertNoKey(t, kv, "runs")
}

func assertHasKey(t *testing.T, kv []any, want string) {
	t.Helper()
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if ok && key == want {
			return
		}
	}
	t.Fatalf("missing key %q in %v", want, kv)
}

func assertNoKey(t *testing.T, kv []any, banned string) {
	t.Helper()
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if ok && key == banned {
			t.Fatalf("unexpected key %q in %v", banned, kv)
		}
	}
}

func assertNoValue(t *testing.T, kv []any, banned any) {
	t.Helper()
	for i := 1; i < len(kv); i += 2 {
		if kv[i] == banned {
			t.Fatalf("unexpected value %v in %v", banned, kv)
		}
	}
}

func ptr[T any](value T) *T {
	return &value
}
