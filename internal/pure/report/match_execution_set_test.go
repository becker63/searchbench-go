package report

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

func TestMatchExecutionSetValidation(t *testing.T) {
	t.Parallel()

	match1 := domain.MatchID("match-1")
	match2 := domain.MatchID("match-2")

	tests := []struct {
		name    string
		items   map[domain.MatchID]string
		order   []domain.MatchID
		wantErr bool
	}{
		{
			name:    "empty order",
			items:   map[domain.MatchID]string{match1: "a"},
			wantErr: true,
		},
		{
			name:    "missing match",
			items:   map[domain.MatchID]string{match1: "a"},
			order:   []domain.MatchID{match1, match2},
			wantErr: true,
		},
		{
			name: "duplicate match id",
			items: map[domain.MatchID]string{
				match1: "a",
				match2: "b",
			},
			order:   []domain.MatchID{match1, match1},
			wantErr: true,
		},
		{
			name: "valid",
			items: map[domain.MatchID]string{
				match1: "a",
				match2: "b",
			},
			order: []domain.MatchID{match2, match1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			set, err := NewMatchExecutionSet(tt.items, tt.order)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got, want := set.Len(), len(tt.order); got != want {
				t.Fatalf("Len() = %d, want %d", got, want)
			}
			if got := set.Values(); len(got) != len(tt.order) {
				t.Fatalf("len(Values()) = %d, want %d", len(got), len(tt.order))
			}
		})
	}
}

func TestMatchExecutionSetItems(t *testing.T) {
	t.Parallel()

	match1 := domain.MatchID("match-1")
	match2 := domain.MatchID("match-2")
	set, err := NewMatchExecutionSet(
		map[domain.MatchID]string{
			match1: "incumbent",
			match2: "challenger",
		},
		[]domain.MatchID{match2, match1},
	)
	if err != nil {
		t.Fatalf("NewMatchExecutionSet() error = %v", err)
	}

	gotIDs := make([]domain.MatchID, 0)
	gotVals := make([]string, 0)
	for matchID, value := range set.Items() {
		gotIDs = append(gotIDs, matchID)
		gotVals = append(gotVals, value)
	}

	wantIDs := []domain.MatchID{match2, match1}
	wantVals := []string{"challenger", "incumbent"}
	for i := range wantIDs {
		if gotIDs[i] != wantIDs[i] {
			t.Fatalf("gotIDs[%d] = %q, want %q", i, gotIDs[i], wantIDs[i])
		}
		if gotVals[i] != wantVals[i] {
			t.Fatalf("gotVals[%d] = %q, want %q", i, gotVals[i], wantVals[i])
		}
	}
}
