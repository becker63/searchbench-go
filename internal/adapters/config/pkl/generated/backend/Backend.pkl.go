// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package backend

import (
	"encoding"
	"fmt"
)

// Execution backend for a policy system. `iterative_context` is the iterative context challenger stack;
// `jcodemunch` is the legacy path; `fake` is a deterministic local stub for tests and offline runs.
type Backend string

const (
	IterativeContext Backend = "iterative_context"
	Jcodemunch       Backend = "jcodemunch"
	Fake             Backend = "fake"
)

// String returns the string representation of Backend
func (rcv Backend) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Backend)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Backend.
func (rcv *Backend) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "iterative_context":
		*rcv = IterativeContext
	case "jcodemunch":
		*rcv = Jcodemunch
	case "fake":
		*rcv = Fake
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Backend`, str)
	}
	return nil
}
