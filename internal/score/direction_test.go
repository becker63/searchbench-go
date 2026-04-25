package score

import "testing"

func TestImproved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		direction Direction
		baseline  float64
		candidate float64
		want      bool
	}{
		{
			name:      "higher is better improves",
			direction: HigherIsBetter,
			baseline:  0.5,
			candidate: 0.8,
			want:      true,
		},
		{
			name:      "higher is better regression",
			direction: HigherIsBetter,
			baseline:  0.8,
			candidate: 0.5,
			want:      false,
		},
		{
			name:      "lower is better improves",
			direction: LowerIsBetter,
			baseline:  5,
			candidate: 3,
			want:      true,
		},
		{
			name:      "lower is better regression",
			direction: LowerIsBetter,
			baseline:  3,
			candidate: 5,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Improved(tt.direction, tt.baseline, tt.candidate); got != tt.want {
				t.Fatalf("Improved(%q, %v, %v) = %v, want %v", tt.direction, tt.baseline, tt.candidate, got, tt.want)
			}
		})
	}
}
