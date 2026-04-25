package backend

import "context"

// Backend may be shared by a runner. Implementations that support parallel
// execution must make StartSession safe for concurrent calls or must be wrapped
// by a sequential runner configuration.
type Backend interface {
	StartSession(ctx context.Context, spec SessionSpec) (Session, error)
}
