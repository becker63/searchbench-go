package compare

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
)

type ExecutionMode string

const (
	// ExecutionSequential compares tasks one at a time in plan order.
	ExecutionSequential ExecutionMode = "sequential"
	// ExecutionParallel compares tasks concurrently while preserving report order.
	ExecutionParallel ExecutionMode = "parallel"
)

// Parallelism configures task-level orchestration policy for Runner.
//
// It does not introduce concurrency into the domain model itself. Only task
// execution in compare.RunTasks may run concurrently.
type Parallelism struct {
	Mode       ExecutionMode `json:"mode"`
	MaxWorkers int           `json:"max_workers"`
	FailFast   bool          `json:"fail_fast"`
}

// TaskWorkResult captures one completed task comparison together with its
// original task index.
//
// Runner uses the index to restore plan order after parallel work completes.
type TaskWorkResult struct {
	Index  int
	TaskID domain.TaskID
	Result TaskComparisonResult
}

// DefaultParallelism returns the default runner policy: sequential task
// execution with a single worker.
func DefaultParallelism() Parallelism {
	return Parallelism{
		Mode:       ExecutionSequential,
		MaxWorkers: 1,
	}
}

// Normalize fills in default values while preserving the requested execution
// mode semantics.
//
// Sequential execution always normalizes to one worker. Parallel execution
// preserves the caller's worker count so validation can reject invalid values.
func (p Parallelism) Normalize() Parallelism {
	if p.Mode == "" {
		p.Mode = ExecutionSequential
	}

	if p.Mode == ExecutionSequential {
		p.MaxWorkers = 1
	}

	return p
}

// Validate checks that the requested execution policy is internally coherent.
func (p Parallelism) Validate() error {
	switch p.Mode {
	case "", ExecutionSequential:
		return nil
	case ExecutionParallel:
		if p.MaxWorkers <= 0 {
			return fmt.Errorf("compare: parallel execution requires max_workers > 0")
		}
		return nil
	default:
		return fmt.Errorf("compare: unknown execution mode %q", p.Mode)
	}
}
