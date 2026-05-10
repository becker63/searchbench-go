// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package runmode

import (
	"encoding"
	"fmt"
)

type RunMode string

const (
	Evaluation   RunMode = "evaluation"
	Optimization RunMode = "optimization"
)

// String returns the string representation of RunMode
func (rcv RunMode) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(RunMode)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for RunMode.
func (rcv *RunMode) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "evaluation":
		*rcv = Evaluation
	case "optimization":
		*rcv = Optimization
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid RunMode`, str)
	}
	return nil
}
