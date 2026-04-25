package run

import "github.com/becker63/searchbench-go/internal/domain"

type FailureStage string

const (
	// FailurePrepare identifies prepare/session setup failures.
	FailurePrepare FailureStage = "prepare"
	// FailureExecute identifies execution-time failures.
	FailureExecute FailureStage = "execute"
	// FailureScore identifies scoring-time failures.
	FailureScore FailureStage = "score"
	// FailureReport identifies report construction failures.
	FailureReport FailureStage = "report"
)

// RunFailure records the report-facing failed path for one task/system run.
//
// ExecutedRun means execution succeeded. score.ScoredRun means scoring
// succeeded. RunFailure is the separate artifact used when one of those steps
// could not complete.
type RunFailure struct {
	RunID   domain.RunID    `json:"run_id"`
	TaskID  domain.TaskID   `json:"task_id"`
	System  domain.SystemID `json:"system"`
	Stage   FailureStage    `json:"stage"`
	Message string          `json:"message"`
}

// NewFailure constructs a report-facing failure record from a run spec.
func NewFailure(spec Spec, stage FailureStage, message string) RunFailure {
	return RunFailure{
		RunID:   spec.ID,
		TaskID:  spec.Task.ID,
		System:  spec.System.ID,
		Stage:   stage,
		Message: message,
	}
}
