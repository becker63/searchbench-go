package report

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/domain"
)

func TestNewTaskRunSetValidation(t *testing.T) {
	t.Parallel()

	task1 := domain.TaskID("task-1")
	task2 := domain.TaskID("task-2")

	tests := []struct {
		name    string
		items   map[domain.TaskID]string
		order   []domain.TaskID
		wantErr bool
	}{
		{
			name:    "empty order",
			items:   map[domain.TaskID]string{task1: "a"},
			wantErr: true,
		},
		{
			name:    "missing task",
			items:   map[domain.TaskID]string{task1: "a"},
			order:   []domain.TaskID{task1, task2},
			wantErr: true,
		},
		{
			name: "duplicate task id",
			items: map[domain.TaskID]string{
				task1: "a",
				task2: "b",
			},
			order:   []domain.TaskID{task1, task1},
			wantErr: true,
		},
		{
			name: "valid",
			items: map[domain.TaskID]string{
				task1: "a",
				task2: "b",
			},
			order: []domain.TaskID{task2, task1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			set, err := NewTaskRunSet(tt.items, tt.order)
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
