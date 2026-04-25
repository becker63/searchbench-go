package score

import "testing"

func TestScoreSetPoints(t *testing.T) {
	t.Parallel()

	set := ScoreSet{
		GoldHop:         Metric[HopDistance]{Name: MetricGoldHop, Value: 1},
		IssueHop:        Metric[HopDistance]{Name: MetricIssueHop, Value: 2},
		TokenEfficiency: Metric[EfficiencyScore]{Name: MetricTokenEfficiency, Value: 0.7},
		Cost:            Metric[CostScore]{Name: MetricCost, Value: 0.2},
		Composite:       Metric[CompositeScore]{Name: MetricComposite, Value: 0.9},
	}

	got := make([]MetricPoint, 0)
	for point := range set.Points() {
		got = append(got, point)
	}

	wantNames := []MetricName{
		MetricGoldHop,
		MetricIssueHop,
		MetricTokenEfficiency,
		MetricCost,
		MetricComposite,
	}
	wantDirections := []Direction{
		LowerIsBetter,
		LowerIsBetter,
		HigherIsBetter,
		LowerIsBetter,
		HigherIsBetter,
	}

	if len(got) != len(wantNames) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(wantNames))
	}
	for i := range wantNames {
		if got[i].Name != wantNames[i] {
			t.Fatalf("got[%d].Name = %q, want %q", i, got[i].Name, wantNames[i])
		}
		if got[i].Direction != wantDirections[i] {
			t.Fatalf("got[%d].Direction = %q, want %q", i, got[i].Direction, wantDirections[i])
		}
	}
}
