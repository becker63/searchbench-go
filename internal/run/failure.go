package run

import "github.com/becker63/searchbench-go/internal/domain"

type FailureStage string

const (
	FailurePrepare FailureStage = "prepare"
	FailureExecute FailureStage = "execute"
	FailureScore   FailureStage = "score"
	FailureReport  FailureStage = "report"
)

// RunFailure records a failed path for one task/system run.
type RunFailure struct {
	RunID   domain.RunID    `json:"run_id"`
	TaskID  domain.TaskID   `json:"task_id"`
	System  domain.SystemID `json:"system"`
	Stage   FailureStage    `json:"stage"`
	Message string          `json:"message"`
}

func NewFailure(spec Spec, stage FailureStage, message string) RunFailure {
	return RunFailure{
		RunID:   spec.ID,
		TaskID:  spec.Task.ID,
		System:  spec.System.ID,
		Stage:   stage,
		Message: message,
	}
}
