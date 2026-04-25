package run

import "github.com/becker63/searchbench-go/internal/domain"

// Spec is a planned request to run one system on one task.
//
// Conceptually:
//
//	RunSpec = TaskSpec + SystemSpec
//
// It should be deterministic and serializable. Runtime state belongs in
// PreparedRun or ExecutedRun, not here.
type Spec struct {
	ID     domain.RunID      `json:"id"`
	Task   domain.TaskSpec   `json:"task"`
	System domain.SystemSpec `json:"system"`
}

func NewSpec(id domain.RunID, task domain.TaskSpec, system domain.SystemSpec) Spec {
	return Spec{
		ID:     id,
		Task:   task,
		System: system,
	}
}
