package backend

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
)

type Operation string

const (
	OperationStartSession Operation = "start_session"
	OperationListTools    Operation = "list_tools"
	OperationCallTool     Operation = "call_tool"
	OperationCloseSession Operation = "close_session"
)

type Error struct {
	Backend domain.BackendKind
	Op      Operation
	Err     error
}

func (e Error) Error() string {
	return fmt.Sprintf("backend=%s op=%s: %v", e.Backend, e.Op, e.Err)
}

func (e Error) Unwrap() error {
	return e.Err
}
