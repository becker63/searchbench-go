package round

import (
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestEnrichLocalizationEvidenceCountsPredictionMisses(t *testing.T) {
	t.Parallel()
	gold := domain.RepoRelPath("pkg/gold.go")
	task := domain.MatchSpec{
		ID:        domain.MatchID("m1"),
		Benchmark: domain.BenchmarkLCA,
		Oracle:    domain.MatchOracle{GoldFiles: []domain.RepoRelPath{gold}},
	}
	inc := scoredRunForLocalizationTestRole(t, task, domain.RoleIncumbent, domain.Prediction{Files: []domain.RepoRelPath{"pkg/wrong.go"}})
	ch := scoredRunForLocalizationTestRole(t, task, domain.RoleChallenger, domain.Prediction{Files: []domain.RepoRelPath{gold}})
	rec := []report.MatchExecutionRecord{{
		MatchID:    task.ID,
		Task:       task,
		Incumbent:  report.NewRoleExecutionOutcome(&inc, nil),
		Challenger: report.NewRoleExecutionOutcome(&ch, nil),
	}}
	ev := score.RoundEvidenceDocument{SchemaVersion: score.EvidenceSchemaVersion, ReportID: domain.ReportID("r1")}
	enrichLocalizationEvidence(&ev, rec)
	if !ev.InvalidPredictions.Known || ev.InvalidPredictions.Count != 1 {
		t.Fatalf("InvalidPredictions = %#v, want known:true count:1", ev.InvalidPredictions)
	}
}

func TestProjectRoundEvidenceLeavesUsageUnavailableWhenRunsOmitUsage(t *testing.T) {
	t.Parallel()
	gold := domain.RepoRelPath("pkg/gold.go")
	task := domain.MatchSpec{
		ID:        domain.MatchID("m1"),
		Benchmark: domain.BenchmarkLCA,
		Oracle:    domain.MatchOracle{GoldFiles: []domain.RepoRelPath{gold}},
	}
	incSys := domain.SystemSpec{
		ID:           domain.SystemID("inc"),
		Name:         "Inc",
		Backend:      domain.BackendFake,
		Model:        domain.ModelSpec{Provider: "fake", Name: "fake"},
		PromptBundle: domain.PromptBundleRef{Name: "b"},
	}
	chSys := domain.SystemSpec{
		ID:           domain.SystemID("ch"),
		Name:         "Ch",
		Backend:      domain.BackendFake,
		Model:        domain.ModelSpec{Provider: "fake", Name: "fake"},
		PromptBundle: domain.PromptBundleRef{Name: "b"},
	}
	inc := scoredRunForLocalizationTestNoUsage(t, task, incSys, domain.RoleIncumbent, domain.Prediction{Files: []domain.RepoRelPath{gold}})
	ch := scoredRunForLocalizationTestNoUsage(t, task, chSys, domain.RoleChallenger, domain.Prediction{Files: []domain.RepoRelPath{gold}})
	tasks, err := domain.NonEmptyFromSlice([]domain.MatchSpec{task})
	if err != nil {
		t.Fatal(err)
	}
	spec := report.NewComparisonSpec(domain.NewPair(incSys, chSys), tasks)
	rr := report.NewRoundReport(
		domain.ReportID("report-usage-test"),
		spec,
		domain.NewPair([]score.ScoredRun{inc}, []score.ScoredRun{ch}),
		domain.NewPair([]run.RunFailure{}, []run.RunFailure{}),
		[]report.ScoreComparison{
			report.NewScoreComparison(score.MetricComposite, 0.2, 0.8),
		},
		nil,
		report.Decision{Decision: report.DecisionReview, Reason: "test"},
	)
	rr.CreatedAt = time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)

	plan := Plan{
		Game:     GameConfig{ID: "game"},
		Round:    RoundConfig{ID: "round"},
		ReportID: domain.ReportID("report-usage-test"),
	}
	got, err := projectRoundEvidence(plan, rr, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.ChallengerUsage.Available || got.IncumbentUsage.Available {
		t.Fatalf("usage should stay unavailable when modeled usage summaries are empty: challenger=%#v incumbent=%#v",
			got.ChallengerUsage, got.IncumbentUsage)
	}
	if got.ChallengerUsage.TotalTokens != 0 || got.IncumbentUsage.TotalTokens != 0 {
		t.Fatalf("token totals should not invent zeros as measurements: challenger=%#v incumbent=%#v",
			got.ChallengerUsage, got.IncumbentUsage)
	}
}

func scoredRunForLocalizationTestRole(t *testing.T, task domain.MatchSpec, role domain.Role, pred domain.Prediction) score.ScoredRun {
	t.Helper()
	sys := domain.SystemSpec{
		ID:           domain.SystemID("sys"),
		Name:         "Sys",
		Backend:      domain.BackendFake,
		Model:        domain.ModelSpec{Provider: "fake", Name: "fake"},
		PromptBundle: domain.PromptBundleRef{Name: "b"},
	}
	return scoredRunForLocalizationTestNoUsage(t, task, sys, role, pred)
}

func scoredRunForLocalizationTestNoUsage(t *testing.T, task domain.MatchSpec, sys domain.SystemSpec, role domain.Role, pred domain.Prediction) score.ScoredRun {
	t.Helper()
	spec := run.NewSpec(domain.RunID(string(role)+"-"+task.ID.String()), task, sys)
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("session-"+task.ID.String()))
	executed := run.NewExecuted(
		prepared,
		pred,
		domain.UsageSummary{},
		domain.TraceID("trace-"+task.ID.String()),
		time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 12, 12, 0, 1, 0, time.UTC),
	)
	ss, err := score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 1},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 1},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.5},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.5},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.5},
	)
	if err != nil {
		t.Fatal(err)
	}
	sr, err := score.NewScoredRun(executed, ss)
	if err != nil {
		t.Fatal(err)
	}
	return sr
}
