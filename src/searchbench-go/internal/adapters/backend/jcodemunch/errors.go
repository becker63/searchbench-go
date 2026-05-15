package jcodemunch

import (
	"errors"
	"fmt"
)

// Kind classifies jCodeMunch adapter failures for errors.Is-style checks.
type Kind int

const (
	// KindSetup covers MCP session setup: connect, initialize, optional repo
	// indexing hooks, and prefetch steps such as tools/list when building the tool surface.
	KindSetup Kind = iota
	// KindToolCall covers MCP tools/call failures and tool results marked as errors.
	KindToolCall
)

// Error is a typed adapter failure with an operation label.
type Error struct {
	Kind Kind
	Op   string
	Err  error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	kind := "setup"
	if e.Kind == KindToolCall {
		kind = "tool"
	}
	if e.Err == nil {
		return fmt.Sprintf("jcodemunch %s: %s", kind, e.Op)
	}
	return fmt.Sprintf("jcodemunch %s: %s: %v", kind, e.Op, e.Err)
}

// Unwrap implements error unwrapper semantics.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// IsSetup reports whether err is (or wraps) a [Error] with [KindSetup].
func IsSetup(err error) bool {
	var je *Error
	return errors.As(err, &je) && je != nil && je.Kind == KindSetup
}

// IsToolCall reports whether err is (or wraps) a [Error] with [KindToolCall].
func IsToolCall(err error) bool {
	var je *Error
	return errors.As(err, &je) && je != nil && je.Kind == KindToolCall
}
