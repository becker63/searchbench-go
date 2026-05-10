package score

import "testing"

func TestImproved(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		direction  Direction
		incumbent  float64
		challenger float64
		want       bool
	}{
		{
			name:       "higher is better improves",
			direction:  HigherIsBetter,
			incumbent:  0.5,
			challenger: 0.8,
			want:       true,
		},
		{
			name:       "higher is better regression",
			direction:  HigherIsBetter,
			incumbent:  0.8,
			challenger: 0.5,
			want:       false,
		},
		{
			name:       "lower is better improves",
			direction:  LowerIsBetter,
			incumbent:  5,
			challenger: 3,
			want:       true,
		},
		{
			name:       "lower is better regression",
			direction:  LowerIsBetter,
			incumbent:  3,
			challenger: 5,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Improved(tt.direction, tt.incumbent, tt.challenger); got != tt.want {
				t.Fatalf("Improved(%q, %v, %v) = %v, want %v", tt.direction, tt.incumbent, tt.challenger, got, tt.want)
			}
		})
	}
}
