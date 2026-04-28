package domain

import "testing"

func TestPairAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		pair      Pair[string]
		wantRoles []Role
		wantVals  []string
		stopAfter int
	}{
		{
			name:      "yields both roles in order",
			pair:      NewPair("baseline", "candidate"),
			wantRoles: []Role{RoleBaseline, RoleCandidate},
			wantVals:  []string{"baseline", "candidate"},
			stopAfter: 2,
		},
		{
			name:      "respects early stop",
			pair:      NewPair("baseline", "candidate"),
			wantRoles: []Role{RoleBaseline},
			wantVals:  []string{"baseline"},
			stopAfter: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotRoles := make([]Role, 0)
			gotVals := make([]string, 0)
			count := 0
			for role, value := range tt.pair.All() {
				gotRoles = append(gotRoles, role)
				gotVals = append(gotVals, value)
				count++
				if count >= tt.stopAfter {
					break
				}
			}

			if len(gotRoles) != len(tt.wantRoles) {
				t.Fatalf("len(gotRoles) = %d, want %d", len(gotRoles), len(tt.wantRoles))
			}
			for i := range tt.wantRoles {
				if gotRoles[i] != tt.wantRoles[i] {
					t.Fatalf("gotRoles[%d] = %q, want %q", i, gotRoles[i], tt.wantRoles[i])
				}
				if gotVals[i] != tt.wantVals[i] {
					t.Fatalf("gotVals[%d] = %q, want %q", i, gotVals[i], tt.wantVals[i])
				}
			}
		})
	}
}
