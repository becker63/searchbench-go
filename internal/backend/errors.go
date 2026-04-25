package backend

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
)

type Operation string

const (
	// OperationStartSession identifies backend session creation.
	OperationStartSession Operation = "start_session"
	// OperationListTools identifies backend tool discovery.
	OperationListTools Operation = "list_tools"
	// OperationCallTool identifies one backend tool invocation.
	OperationCallTool Operation = "call_tool"
	// OperationCloseSession identifies backend session teardown.
	OperationCloseSession Operation = "close_session"
)

// Error classifies backend failures by backend kind and operation.
type Error struct {
	Backend domain.BackendKind
	Op      Operation
	Err     error
}

// Error formats the backend operation plus the wrapped cause.
func (e Error) Error() string {
	return fmt.Sprintf("backend=%s op=%s: %v", e.Backend, e.Op, e.Err)
}

// Unwrap returns the underlying backend cause.
func (e Error) Unwrap() error {
	return e.Err
}
