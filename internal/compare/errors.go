package compare

import (
	"errors"
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/run"
)

type StageError struct {
	Stage    run.FailureStage
	RunID    domain.RunID
	TaskID   domain.TaskID
	SystemID domain.SystemID
	Err      error
}

func NewStageError(spec run.Spec, stage run.FailureStage, err error) StageError {
	if err == nil {
		err = errors.New("unknown error")
	}
	return StageError{
		Stage:    stage,
		RunID:    spec.ID,
		TaskID:   spec.Task.ID,
		SystemID: spec.System.ID,
		Err:      err,
	}
}

func (e StageError) Error() string {
	return fmt.Sprintf(
		"stage=%s run_id=%s task_id=%s system_id=%s: %v",
		e.Stage,
		e.RunID,
		e.TaskID,
		e.SystemID,
		e.Err,
	)
}

func (e StageError) Unwrap() error {
	return e.Err
}

func failureFromError(spec run.Spec, stage run.FailureStage, err error) run.RunFailure {
	var stageErr StageError
	if errors.As(err, &stageErr) {
		return run.RunFailure{
			RunID:   stageErr.RunID,
			TaskID:  stageErr.TaskID,
			System:  stageErr.SystemID,
			Stage:   stageErr.Stage,
			Message: err.Error(),
		}
	}

	if err == nil {
		err = errors.New("unknown error")
	}
	return run.NewFailure(spec, stage, err.Error())
}
