package gitmaterialize

import (
	"errors"
	"fmt"
)

// FailureKind classifies git materialization faults.
type FailureKind string

const (
	FailureInvalidRepoURL  FailureKind = "invalid_repo_url"
	FailureCloneFailed     FailureKind = "clone_failed"
	FailureFetchFailed     FailureKind = "fetch_failed"
	FailureCheckoutFailed  FailureKind = "checkout_failed"
	FailureMissingBaseSHA  FailureKind = "missing_base_sha"
	FailureCachePermission FailureKind = "cache_permission_failed"
	FailureInvalidTaskRepo FailureKind = "invalid_task_repo"
	FailureFilesystemError FailureKind = "filesystem_error"
)

// MaterializeError wraps a typed materialization failure.
type MaterializeError struct {
	Kind FailureKind
	Err  error
}

func (e *MaterializeError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("materialize %s: %v", e.Kind, e.Err)
	}
	return fmt.Sprintf("materialize %s", e.Kind)
}

func (e *MaterializeError) Unwrap() error { return e.Err }

// AsFailureKind unwraps Kind from a [*MaterializeError], if any.
func AsFailureKind(err error) (FailureKind, bool) {
	var me *MaterializeError
	if errors.As(err, &me) {
		return me.Kind, true
	}
	return "", false
}

func newErr(kind FailureKind, err error) error {
	return &MaterializeError{Kind: kind, Err: err}
}
