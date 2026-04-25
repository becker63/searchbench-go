package domain

import (
	"encoding/json"
	"testing"
)

func TestNonEmptyFromSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		items   []int
		wantErr bool
	}{
		{
			name:    "empty",
			items:   nil,
			wantErr: true,
		},
		{
			name:  "single item",
			items: []int{7},
		},
		{
			name:  "multiple items",
			items: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NonEmptyFromSlice(tt.items)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			all := got.All()
			if len(all) != len(tt.items) {
				t.Fatalf("len(all) = %d, want %d", len(all), len(tt.items))
			}
			for i := range all {
				if all[i] != tt.items[i] {
					t.Fatalf("all[%d] = %d, want %d", i, all[i], tt.items[i])
				}
			}

			if len(tt.items) > 1 {
				tail := got.Tail()
				tail[0] = 999
				if got.Tail()[0] == 999 {
					t.Fatal("Tail exposed internal slice")
				}
			}
		})
	}
}

func TestNonEmptyZeroValue(t *testing.T) {
	t.Parallel()

	var n NonEmpty[int]

	if n.Valid() {
		t.Fatal("zero-value NonEmpty should be invalid")
	}
	if got := n.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
	if got := n.All(); got != nil {
		t.Fatalf("All() = %#v, want nil", got)
	}
	if got := n.Tail(); got != nil {
		t.Fatalf("Tail() = %#v, want nil", got)
	}
	if err := n.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
	if _, err := json.Marshal(n); err == nil {
		t.Fatal("expected marshal error")
	}

	var values []int
	for value := range n.Values() {
		values = append(values, value)
	}
	if len(values) != 0 {
		t.Fatalf("len(Values()) = %d, want 0", len(values))
	}
}

func TestNonEmptyValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input NonEmpty[int]
		want  []int
		stop  int
	}{
		{
			name:  "all values",
			input: NewNonEmpty(1, 2, 3),
			want:  []int{1, 2, 3},
			stop:  3,
		},
		{
			name:  "early stop",
			input: NewNonEmpty(4, 5, 6),
			want:  []int{4, 5},
			stop:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := make([]int, 0)
			for value := range tt.input.Values() {
				got = append(got, value)
				if len(got) >= tt.stop {
					break
				}
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}
