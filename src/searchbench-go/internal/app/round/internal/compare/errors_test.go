package compare

import (
	"errors"
	"fmt"
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
)

func TestStageErrorUnwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("boom")
	spec := run.NewSpec(
		domain.RunID("run-1"),
		domain.MatchSpec{ID: domain.MatchID("task-1")},
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
		domain.MatchSpec{ID: domain.MatchID("task-1")},
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
		domain.MatchSpec{ID: domain.MatchID("task-spec")},
		domain.SystemSpec{ID: domain.SystemID("system-spec")},
	)
	stageSpec := run.NewSpec(
		domain.RunID("run-stage"),
		domain.MatchSpec{ID: domain.MatchID("task-stage")},
		domain.SystemSpec{ID: domain.SystemID("system-stage")},
	)
	err := fmt.Errorf("wrapped: %w", NewStageError(stageSpec, run.FailureExecute, errors.New("explode")))

	failure := failureFromError(spec, run.FailureScore, err)
	if failure.RunID != stageSpec.ID {
		t.Fatalf("RunID = %q, want %q", failure.RunID, stageSpec.ID)
	}
	if failure.MatchID != stageSpec.Match.ID {
		t.Fatalf("MatchID = %q, want %q", failure.MatchID, stageSpec.Match.ID)
	}
	if failure.System != stageSpec.System.ID {
		t.Fatalf("System = %q, want %q", failure.System, stageSpec.System.ID)
	}
	if failure.Stage != run.FailureExecute {
		t.Fatalf("Stage = %q, want %q", failure.Stage, run.FailureExecute)
	}
}
