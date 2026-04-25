package compare

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
)

type ExecutionMode string

const (
	ExecutionSequential ExecutionMode = "sequential"
	ExecutionParallel   ExecutionMode = "parallel"
)

type Parallelism struct {
	Mode       ExecutionMode `json:"mode"`
	MaxWorkers int           `json:"max_workers"`
	FailFast   bool          `json:"fail_fast"`
}

type TaskWorkResult struct {
	Index  int
	TaskID domain.TaskID
	Result TaskComparisonResult
}

func DefaultParallelism() Parallelism {
	return Parallelism{
		Mode:       ExecutionSequential,
		MaxWorkers: 1,
	}
}

func (p Parallelism) Normalize() Parallelism {
	if p.Mode == "" {
		p.Mode = ExecutionSequential
	}

	if p.Mode == ExecutionSequential {
		p.MaxWorkers = 1
	}

	return p
}

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
