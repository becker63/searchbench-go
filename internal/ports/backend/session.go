package backend

import (
	"context"
	"encoding/json"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// ToolSpec describes one backend-exposed tool using a small protocol shape.
//
// It is intentionally minimal so future iterative-context, jCodeMunch, or
// other adapters can map their native tool representations into a stable
// Searchbench boundary.
type ToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
}

// ToolResult is the normalized response returned by a backend tool call.
type ToolResult struct {
	Name   string          `json:"name"`
	Output json.RawMessage `json:"output"`
}

// SessionSpec identifies the isolated runtime session to start for one run.
type SessionSpec struct {
	ID     domain.SessionID  `json:"id"`
	Task   domain.TaskSpec   `json:"task"`
	System domain.SystemSpec `json:"system"`
}

// Session represents one isolated run/session. Session implementations are not
// required to be safe for concurrent tool calls unless explicitly documented.
type Session interface {
	ID() domain.SessionID
	Tools(ctx context.Context) ([]ToolSpec, error)
	CallTool(ctx context.Context, name string, args json.RawMessage) (ToolResult, error)
	Close(ctx context.Context) error
}
