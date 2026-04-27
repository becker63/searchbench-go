package artifact

import "fmt"

// FailureKind classifies bundle writer failures.
type FailureKind string

const (
	FailureKindValidationFailed    FailureKind = "bundle_validation_failed"
	FailureKindFilesystemFailed    FailureKind = "bundle_filesystem_failed"
	FailureKindSerializationFailed FailureKind = "bundle_serialization_failed"
	FailureKindAlreadyExists       FailureKind = "bundle_already_exists"
	FailureKindFinalizeFailed      FailureKind = "bundle_finalize_failed"
	FailureKindUnexpectedInternal  FailureKind = "unexpected_internal_failure"
)

// Error is the typed bundle writer failure.
type Error struct {
	Phase string
	Kind  FailureKind
	Path  string
	Err   error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil && e.Path == "" {
		return fmt.Sprintf("%s/%s", e.Phase, e.Kind)
	}
	if e.Err == nil {
		return fmt.Sprintf("%s/%s: %s", e.Phase, e.Kind, e.Path)
	}
	if e.Path == "" {
		return fmt.Sprintf("%s/%s: %v", e.Phase, e.Kind, e.Err)
	}
	return fmt.Sprintf("%s/%s: %s: %v", e.Phase, e.Kind, e.Path, e.Err)
}

// Unwrap exposes the underlying cause.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
