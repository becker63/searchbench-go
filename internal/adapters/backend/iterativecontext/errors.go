package iterativecontext

import (
	"errors"
	"fmt"
)

// Kind classifies Iterative Context adapter failures for errors.Is-style checks.
type Kind int

const (
	// KindSession covers MCP transport/session setup before a usable client session exists
	// (connect, handshake).
	KindSession Kind = iota
	// KindInstall covers harness score installation via MCP install_score.
	KindInstall
	// KindVerify covers harness score verification via MCP verify_score.
	KindVerify
	// KindToolSetup covers evaluator tool surface preparation (tools/list and schema mapping).
	KindToolSetup
	// KindToolCall covers evaluator MCP tools/call failures and tool results marked as errors.
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
	kind := "session"
	switch e.Kind {
	case KindInstall:
		kind = "install"
	case KindVerify:
		kind = "verify"
	case KindToolSetup:
		kind = "tool_setup"
	case KindToolCall:
		kind = "tool"
	}
	if e.Err == nil {
		return fmt.Sprintf("iterativecontext %s: %s", kind, e.Op)
	}
	return fmt.Sprintf("iterativecontext %s: %s: %v", kind, e.Op, e.Err)
}

// Unwrap implements error unwrapper semantics.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// IsSession reports whether err is (or wraps) an [Error] with [KindSession].
func IsSession(err error) bool {
	var ie *Error
	return errors.As(err, &ie) && ie != nil && ie.Kind == KindSession
}

// IsInstall reports whether err is (or wraps) an [Error] with [KindInstall].
func IsInstall(err error) bool {
	var ie *Error
	return errors.As(err, &ie) && ie != nil && ie.Kind == KindInstall
}

// IsVerify reports whether err is (or wraps) an [Error] with [KindVerify].
func IsVerify(err error) bool {
	var ie *Error
	return errors.As(err, &ie) && ie != nil && ie.Kind == KindVerify
}

// IsToolSetup reports whether err is (or wraps) an [Error] with [KindToolSetup].
func IsToolSetup(err error) bool {
	var ie *Error
	return errors.As(err, &ie) && ie != nil && ie.Kind == KindToolSetup
}

// IsToolCall reports whether err is (or wraps) an [Error] with [KindToolCall].
func IsToolCall(err error) bool {
	var ie *Error
	return errors.As(err, &ie) && ie != nil && ie.Kind == KindToolCall
}
