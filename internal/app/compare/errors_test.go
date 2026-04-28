package compare

import (
	"errors"
	"fmt"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/run"
)

func TestStageErrorUnwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("boom")
	spec := run.NewSpec(
		domain.RunID("run-1"),
		domain.TaskSpec{ID: domain.TaskID("task-1")},
		domain.SystemSpec{ID: domain.SystemID("system-1")},
	)

	stageErr := NewStageError(spec, run.FailureExecute, cause)
	if !errors.Is(stageErr, cause) {
		t.Fatal("StageError should unwrap to original cause")
	}
	if got := stageErr.Unwrap(); got != cause {
		t.Fatalf("Unwrap() = %v, want %v", got, cause)
	}
}

func TestStageErrorAsThroughWrapping(t *testing.T) {
	t.Parallel()

	spec := run.NewSpec(
		domain.RunID("run-1"),
		domain.TaskSpec{ID: domain.TaskID("task-1")},
		domain.SystemSpec{ID: domain.SystemID("system-1")},
	)
	err := fmt.Errorf("outer: %w", NewStageError(spec, run.FailureScore, errors.New("bad score")))

	var stageErr StageError
	if !errors.As(err, &stageErr) {
		t.Fatal("errors.As should find wrapped StageError")
	}
	if stageErr.Stage != run.FailureScore {
		t.Fatalf("Stage = %q, want %q", stageErr.Stage, run.FailureScore)
	}
}

func TestFailureFromErrorUsesStageErrorFields(t *testing.T) {
	t.Parallel()

	spec := run.NewSpec(
		domain.RunID("run-spec"),
		domain.TaskSpec{ID: domain.TaskID("task-spec")},
		domain.SystemSpec{ID: domain.SystemID("system-spec")},
	)
	stageSpec := run.NewSpec(
		domain.RunID("run-stage"),
		domain.TaskSpec{ID: domain.TaskID("task-stage")},
		domain.SystemSpec{ID: domain.SystemID("system-stage")},
	)
	err := fmt.Errorf("wrapped: %w", NewStageError(stageSpec, run.FailureExecute, errors.New("explode")))

	failure := failureFromError(spec, run.FailureScore, err)
	if failure.RunID != stageSpec.ID {
		t.Fatalf("RunID = %q, want %q", failure.RunID, stageSpec.ID)
	}
	if failure.TaskID != stageSpec.Task.ID {
		t.Fatalf("TaskID = %q, want %q", failure.TaskID, stageSpec.Task.ID)
	}
	if failure.System != stageSpec.System.ID {
		t.Fatalf("System = %q, want %q", failure.System, stageSpec.System.ID)
	}
	if failure.Stage != run.FailureExecute {
		t.Fatalf("Stage = %q, want %q", failure.Stage, run.FailureExecute)
	}
}
