package eino

import (
	"fmt"
)

// Failure is the typed failed outcome for one evaluator run.
//
// This remains evaluator-local because it carries retry details that are not
// yet shared across executors.
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
