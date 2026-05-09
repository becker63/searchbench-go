package report

import (
	"math"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestRoundEvidenceUsesIncumbentChallengerRoles(t *testing.T) {
	t.Parallel()

	report := sampleRoundReport(t)
	got, err := BuildRoundEvidence(report)
	if err != nil {
		t.Fatalf("BuildRoundEvidence() error = %v", err)
	}

	if got.ReportID != report.ID {
		t.Fatalf("ReportID = %q, want %q", got.ReportID, report.ID)
	}
	if got.ExecutionCounts.Incumbent != 2 || got.ExecutionCounts.Challenger != 2 {
		t.Fatalf("ExecutionCounts = %#v, want 2/2", got.ExecutionCounts)
	}
	if got.FailureCounts.Incumbent != 1 || got.FailureCounts.Challenger != 0 {
		t.Fatalf("FailureCounts = %#v, want 1/0", got.FailureCounts)
	}
	if got.Decision.Decision != string(DecisionReview) {
		t.Fatalf("Decision.Decision = %q, want %q", got.Decision.Decision, DecisionReview)
	}
	if len(got.Metrics) != len(report.Comparisons) {
		t.Fatalf("len(Metrics) = %d, want %d", len(got.Metrics), len(report.Comparisons))
	}
	cost := findMetricEvidence(t, got.Metrics, score.MetricCost)
	if cost.Direction != score.LowerIsBetter {
		t.Fatalf("cost.Direction = %q, want %q", cost.Direction, score.LowerIsBetter)
	}
	if cost.Incumbent != 0.575 || cost.Challenger != 0.15 || math.Abs(cost.Delta-(-0.425)) > 1e-9 {
		t.Fatalf("cost evidence = %#v, want stable values", cost)
	}
	if !cost.Improved || cost.Regressed {
		t.Fatalf("cost flags = improved:%v regressed:%v, want true/false", cost.Improved, cost.Regressed)
	}
	if got.LocalizationDistance.GoldHop == nil || got.LocalizationDistance.IssueHop == nil {
		t.Fatalf("LocalizationDistance = %#v, want gold_hop and issue_hop", got.LocalizationDistance)
	}
	if !got.ChallengerUsage.Available || got.ChallengerUsage.TotalTokens == 0 {
		t.Fatalf("ChallengerUsage = %#v, want aggregated challenger usage", got.ChallengerUsage)
	}
	if got.IncumbentUsage.TotalTokens == 0 {
		t.Fatalf("IncumbentUsage = %#v, want aggregated incumbent usage", got.IncumbentUsage)
	}
	if got.Regressions.Count != 1 || got.Regressions.MinorCount != 1 || got.Regressions.SevereCount != 0 {
		t.Fatalf("Regressions = %#v, want summarized minor regression", got.Regressions)
	}
	if got.InvalidPredictions.Known {
		t.Fatalf("InvalidPredictions = %#v, want unknown in current report substrate", got.InvalidPredictions)
	}
}

func TestBuildRoundEvidenceFailsForMissingReportID(t *testing.T) {
	t.Parallel()

	report := sampleRoundReport(t)
	report.ID = ""

	if _, err := BuildRoundEvidence(report); err == nil {
		t.Fatal("expected error")
	}
}

func findMetricEvidence(t *testing.T, metrics []score.MetricEvidence, name score.MetricName) score.MetricEvidence {
	t.Helper()

	for _, metric := range metrics {
		if metric.Metric == name {
			return metric
		}
	}
	t.Fatalf("metric %q not found", name)
	return score.MetricEvidence{}
}

func sampleRoundReport(t *testing.T) RoundReport {
	t.Helper()

	policySource := "def score(task):\n    return 'candidate'\n"
	baseline := sampleBaselineSystem()
	candidate := sampleCandidateSystem(policySource)
	taskOne := sampleTask(domain.MatchID("task-1"), domain.RepoRelPath("pkg/bug1.go"))
	taskTwo := sampleTask(domain.MatchID("task-2"), domain.RepoRelPath("pkg/bug2.go"))
	tasks := domain.NewNonEmpty(taskOne, taskTwo)
	spec := NewComparisonSpec(domain.NewPair(baseline, candidate), tasks)

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

	out := NewRoundReport(
		domain.ReportID("report-evidence"),
		spec,
		runs,
		failures,
		[]ScoreComparison{
			NewScoreComparison(score.MetricGoldHop, 3.5, 1.5),
			NewScoreComparison(score.MetricIssueHop, 4.5, 1.5),
			NewScoreComparison(score.MetricTokenEfficiency, 0.425, 0.85),
			NewScoreComparison(score.MetricCost, 0.575, 0.15),
			NewScoreComparison(score.MetricComposite, 0.325, 0.90),
		},
		[]Regression{
			{
				MatchID:   taskTwo.ID,
				Metric:    score.MetricCost,
				Baseline:  0.10,
				Candidate: 0.20,
				Delta:     0.10,
				Severity:  RegressionMinor,
				Reason:    "candidate cost is slightly higher on task-2",
			},
		},
		PromotionDecision{
			Decision: DecisionReview,
			Reason:   "candidate improves core metrics but has a minor cost regression",
		},
	)
	out.CreatedAt = time.Date(2026, 4, 26, 15, 4, 5, 0, time.UTC)
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
