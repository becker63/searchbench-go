package compare

import (
	"errors"
	"fmt"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/run"
)

type StageError struct {
	Stage    run.FailureStage
	RunID    domain.RunID
	TaskID   domain.TaskID
	SystemID domain.SystemID
	Err      error
}

// NewStageError constructs a typed orchestration error for one run stage.
//
// StageError is an internal classification helper. Runner eventually converts
// it into the report-facing run.RunFailure artifact once comparison work is
// summarized.
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

// Error formats the stage classification together with the wrapped cause.
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

// Unwrap returns the underlying stage cause for errors.Is/errors.As.
func (e StageError) Unwrap() error {
	return e.Err
}

func failureFromError(spec run.Spec, stage run.FailureStage, err error) run.RunFailure {
	// StageError is the internal classification shape; RunFailure is the
	// report-facing artifact stored in CandidateReport.
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
