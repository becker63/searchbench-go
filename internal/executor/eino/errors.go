package eino

import (
	"fmt"
)

// Failure is the typed failed outcome for one evaluator run.
//
// This remains evaluator-local because it carries retry/attempt details that
// are not yet shared across executors. The failure refers to one evaluator
// attempt or to retry exhaustion across attempts, not to an individual
// Eino-internal model turn.
type Failure struct {
	Phase       Phase
	Kind        FailureKind
	Message     string
	Cause       error
	Recoverable bool
	Attempt     int
}

func (f *Failure) Error() string {
	if f == nil {
		return "<nil>"
	}
	if f.Cause == nil {
		return fmt.Sprintf("%s/%s: %s", f.Phase, f.Kind, f.Message)
	}
	return fmt.Sprintf("%s/%s: %s: %v", f.Phase, f.Kind, f.Message, f.Cause)
}

// Unwrap exposes the underlying cause for errors.As/errors.Is checks.
func (f *Failure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Cause
}
