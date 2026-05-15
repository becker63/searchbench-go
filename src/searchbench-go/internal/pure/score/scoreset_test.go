package score

import "testing"

func TestScoreSetValidate(t *testing.T) {
	t.Parallel()

	valid := ScoreSet{
		GoldHop:         Metric[HopDistance]{Name: MetricGoldHop, Value: 1},
		IssueHop:        Metric[HopDistance]{Name: MetricIssueHop, Value: 2},
		TokenEfficiency: Metric[EfficiencyScore]{Name: MetricTokenEfficiency, Value: 0.8},
		Cost:            Metric[CostScore]{Name: MetricCost, Value: 0.2},
		Composite:       Metric[CompositeScore]{Name: MetricComposite, Value: 0.9},
	}

	tests := []struct {
		name    string
		set     ScoreSet
		wantErr bool
	}{
		{
			name: "valid",
			set:  valid,
		},
		{
			name: "wrong gold hop name",
			set: func() ScoreSet {
				s := valid
				s.GoldHop.Name = MetricComposite
				return s
			}(),
			wantErr: true,
		},
		{
			name: "wrong composite name",
			set: func() ScoreSet {
				s := valid
				s.Composite.Name = MetricCost
				return s
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.set.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
