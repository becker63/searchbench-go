package console

import (
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestRenderRoundReportIncludesHighLevelSections(t *testing.T) {
	t.Parallel()

	out := RenderRoundReport(sampleRoundReport(t), DefaultOptions())

	for _, want := range []string{
		"SearchBench Round Report",
		"Decision",
		"Systems",
		"Execution Summary",
		"Metrics",
		"Regressions",
		"Failures",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n%s", want, out)
		}
	}
}

func TestRenderRoundReportDoesNotLeakPolicySource(t *testing.T) {
	t.Parallel()

	policySource := "def score(task):\n    return 'challenger'\n"
	report := sampleRoundReportWithPolicySource(t, policySource)

	out := RenderRoundReport(report, DefaultOptions())
	if strings.Contains(out, policySource) {
		t.Fatalf("output leaked policy source\n%s", out)
	}
	if strings.Contains(out, "\"source\"") {
		t.Fatalf("output leaked source key\n%s", out)
	}
}

func TestRenderRoundReportShowsFailures(t *testing.T) {
	t.Parallel()

	report := sampleRoundReport(t)
	report.Failures.Challenger = []run.RunFailure{
		{
			RunID:   domain.RunID("challenger-run-1"),
			MatchID: domain.MatchID("task-1"),
			System:  domain.SystemID("challenger-system"),
			Stage:   run.FailureExecute,
			Message: "challenger execute failed",
		},
	}

	out := RenderRoundReport(report, DefaultOptions())
	for _, want := range []string{"Failures", "challenger", "execute", "challenger execute failed"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n%s", want, out)
		}
	}
}

func sampleRoundReport(t *testing.T) report.RoundReport {
	t.Helper()

	return sampleRoundReportWithPolicySource(t, "def score(task):\n    return 'challenger'\n")
}

func sampleRoundReportWithPolicySource(t *testing.T, policySource string) report.RoundReport {
	t.Helper()

	spec := report.NewComparisonSpec(
		domain.NewPair(sampleIncumbentPolicy(), sampleChallengerPolicy(policySource)),
		domain.NewNonEmpty(sampleTask(domain.MatchID("task-1"), domain.RepoRelPath("pkg/bug1.go"))),
	)

	runs := domain.NewPair(
		[]score.ScoredRun{sampleScoredRun(t, domain.RoleIncumbent, sampleIncumbentPolicy(), domain.MatchID("task-1"))},
		[]score.ScoredRun{sampleScoredRun(t, domain.RoleChallenger, sampleChallengerPolicy(policySource), domain.MatchID("task-1"))},
	)
	comparisons := []report.ScoreComparison{
		report.NewScoreComparison(score.MetricGoldHop, 4, 1),
		report.NewScoreComparison(score.MetricIssueHop, 5, 1),
		report.NewScoreComparison(score.MetricTokenEfficiency, 0.4, 0.9),
		report.NewScoreComparison(score.MetricCost, 0.6, 0.1),
		report.NewScoreComparison(score.MetricComposite, 0.3, 0.95),
	}

	out := report.NewRoundReport(
		domain.ReportID("report-1"),
		spec,
		runs,
		domain.NewPair([]run.RunFailure{}, []run.RunFailure{}),
		comparisons,
		nil,
		report.Decision{
			Decision: report.DecisionPromoteChallenger,
			Reason:   "challenger improved every required metric",
		},
	)
	out.CreatedAt = time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	return out
}

func sampleIncumbentPolicy() domain.SystemSpec {
	return domain.SystemSpec{
		ID:      domain.SystemID("incumbent-system"),
		Name:    "Incumbent",
		Backend: domain.BackendJCodeMunch,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-incumbent",
		},
		PromptBundle: domain.PromptBundleRef{
			Name:    "bundle",
			Version: "v1",
		},
	}
}

func sampleChallengerPolicy(policySource string) domain.SystemSpec {
	policy := domain.NewPythonPolicy(domain.PolicyID("policy-1"), policySource, "score")
	return domain.SystemSpec{
		ID:      domain.SystemID("challenger-system"),
		Name:    "Challenger",
		Backend: domain.BackendIterativeContext,
		Model: domain.ModelSpec{
			Provider: "openai",
			Name:     "gpt-challenger",
		},
		PromptBundle: domain.PromptBundleRef{
			Name:    "bundle",
			Version: "v2",
		},
		Policy: &policy,
	}
}

func sampleTask(id domain.MatchID, gold domain.RepoRelPath) domain.MatchSpec {
	return domain.MatchSpec{
		ID:        id,
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("/tmp/repo"),
		},
		Oracle: domain.MatchOracle{
			GoldFiles: []domain.RepoRelPath{gold},
		},
	}
}

func sampleScoredRun(t *testing.T, role domain.Role, system domain.SystemSpec, taskID domain.MatchID) score.ScoredRun {
	t.Helper()

	task := sampleTask(taskID, domain.RepoRelPath("pkg/bug1.go"))
	spec := run.NewSpec(domain.RunID(string(role)+"-"+taskID.String()), task, system)
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("session-"+taskID.String()))
	executed := run.NewExecuted(
		prepared,
		domain.Prediction{Files: []domain.RepoRelPath{"pkg/out.go"}},
		domain.UsageSummary{InputTokens: 10, OutputTokens: 5, TotalTokens: 15, CostUSD: 0.01},
		domain.TraceID("trace-"+taskID.String()),
		time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 24, 12, 0, 2, 0, time.UTC),
	)

	scores, err := score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 1},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 1},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.9},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.1},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.95},
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
