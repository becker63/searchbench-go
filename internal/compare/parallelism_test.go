package compare

import "testing"

func TestParallelismNormalizeAndValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   Parallelism
		want    Parallelism
		wantErr bool
	}{
		{
			name:  "zero value normalizes to sequential",
			input: Parallelism{},
			want: Parallelism{
				Mode:       ExecutionSequential,
				MaxWorkers: 1,
			},
		},
		{
			name: "sequential ignores high max workers",
			input: Parallelism{
				Mode:       ExecutionSequential,
				MaxWorkers: 99,
				FailFast:   true,
			},
			want: Parallelism{
				Mode:       ExecutionSequential,
				MaxWorkers: 1,
				FailFast:   true,
			},
		},
		{
			name: "parallel with workers validates",
			input: Parallelism{
				Mode:       ExecutionParallel,
				MaxWorkers: 2,
			},
			want: Parallelism{
				Mode:       ExecutionParallel,
				MaxWorkers: 2,
			},
		},
		{
			name: "parallel zero workers fails validation",
			input: Parallelism{
				Mode: ExecutionParallel,
			},
			wantErr: true,
		},
		{
			name: "unknown mode fails validation",
			input: Parallelism{
				Mode: ExecutionMode("weird"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			normalized := tt.input.Normalize()
			if tt.wantErr {
				if err := tt.input.Validate(); err == nil {
					t.Fatal("expected validation error")
				}
				return
			}

			if err := normalized.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if normalized != tt.want {
				t.Fatalf("Normalize() = %#v, want %#v", normalized, tt.want)
			}
		})
	}
}
