package console

import (
	"strings"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

func TestRenderCandidateReportIncludesHighLevelSections(t *testing.T) {
	t.Parallel()

	out := RenderCandidateReport(sampleCandidateReport(t), DefaultOptions())

	for _, want := range []string{
		"Searchbench Candidate Report",
		"Decision",
		"Systems",
		"Run Summary",
		"Metrics",
		"Regressions",
		"Failures",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n%s", want, out)
		}
	}
}

func TestRenderCandidateReportDoesNotLeakPolicySource(t *testing.T) {
	t.Parallel()

	policySource := "def score(task):\n    return 'candidate'\n"
	report := sampleCandidateReportWithPolicySource(t, policySource)

	out := RenderCandidateReport(report, DefaultOptions())
	if strings.Contains(out, policySource) {
		t.Fatalf("output leaked policy source\n%s", out)
	}
	if strings.Contains(out, "\"source\"") {
		t.Fatalf("output leaked source key\n%s", out)
	}
}

func TestRenderCandidateReportShowsFailures(t *testing.T) {
	t.Parallel()

	report := sampleCandidateReport(t)
	report.Failures.Candidate = []run.RunFailure{
		{
			RunID:   domain.RunID("candidate-run-1"),
			TaskID:  domain.TaskID("task-1"),
			System:  domain.SystemID("candidate-system"),
			Stage:   run.FailureExecute,
			Message: "candidate execute failed",
		},
	}

	out := RenderCandidateReport(report, DefaultOptions())
	for _, want := range []string{"Failures", "candidate", "execute", "candidate execute failed"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n%s", want, out)
		}
	}
}

func sampleCandidateReport(t *testing.T) report.CandidateReport {
	t.Helper()

	return sampleCandidateReportWithPolicySource(t, "def score(task):\n    return 'candidate'\n")
}

func sampleCandidateReportWithPolicySource(t *testing.T, policySource string) report.CandidateReport {
	t.Helper()

	spec := report.NewComparisonSpec(
		domain.NewPair(sampleBaselineSystem(), sampleCandidateSystem(policySource)),
		domain.NewNonEmpty(sampleTask(domain.TaskID("task-1"), domain.RepoRelPath("pkg/bug1.go"))),
	)

	runs := domain.NewPair(
		[]score.ScoredRun{sampleScoredRun(t, domain.RoleBaseline, sampleBaselineSystem(), domain.TaskID("task-1"))},
		[]score.ScoredRun{sampleScoredRun(t, domain.RoleCandidate, sampleCandidateSystem(policySource), domain.TaskID("task-1"))},
	)
	comparisons := []report.ScoreComparison{
		report.NewScoreComparison(score.MetricGoldHop, 4, 1),
		report.NewScoreComparison(score.MetricIssueHop, 5, 1),
		report.NewScoreComparison(score.MetricTokenEfficiency, 0.4, 0.9),
		report.NewScoreComparison(score.MetricCost, 0.6, 0.1),
		report.NewScoreComparison(score.MetricComposite, 0.3, 0.95),
	}

	out := report.NewCandidateReport(
		domain.ReportID("report-1"),
		spec,
		runs,
		domain.NewPair([]run.RunFailure{}, []run.RunFailure{}),
		comparisons,
		nil,
		report.PromotionDecision{
			Decision: report.DecisionPromote,
			Reason:   "candidate improved every required metric",
		},
	)
	out.CreatedAt = time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	return out
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
	}
}

func sampleTask(id domain.TaskID, gold domain.RepoRelPath) domain.TaskSpec {
	return domain.TaskSpec{
		ID:        id,
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("repo/example"),
			SHA:  domain.RepoSHA("abc123"),
			Path: domain.HostPath("/tmp/repo"),
		},
		Oracle: domain.TaskOracle{
			GoldFiles: []domain.RepoRelPath{gold},
		},
	}
}

func sampleScoredRun(t *testing.T, role domain.Role, system domain.SystemSpec, taskID domain.TaskID) score.ScoredRun {
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
