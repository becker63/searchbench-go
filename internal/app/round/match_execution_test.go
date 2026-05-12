package round

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func TestMatchExecutionRecordsFromTaskWorkPreservesOrder(t *testing.T) {
	t.Parallel()
	a := domain.MatchSpec{ID: domain.MatchID("a")}
	b := domain.MatchSpec{ID: domain.MatchID("b")}
	matches, err := domain.NonEmptyFromSlice([]domain.MatchSpec{a, b})
	if err != nil {
		t.Fatal(err)
	}
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
	ex := run.ExecutedRun{}
	sr := score.ScoredRun{Execution: ex, Scores: ss}

	work := []compare.TaskWorkResult{
		{Index: 1, MatchID: b.ID, Result: compare.TaskComparisonResult{
			Runs: domain.NewPair(&sr, (*score.ScoredRun)(nil)),
		}},
		{Index: 0, MatchID: a.ID, Result: compare.TaskComparisonResult{
			Runs: domain.NewPair((*score.ScoredRun)(nil), &sr),
		}},
	}
	got, err := matchExecutionRecordsFromTaskWork(matches, work)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len %d", len(got))
	}
	if got[0].MatchID != a.ID || got[1].MatchID != b.ID {
		t.Fatalf("order wrong: %#v", got)
	}
	if got[0].Challenger.Scored == nil || got[1].Incumbent.Scored == nil {
		t.Fatal("expected scored pointers preserved")
	}
}
