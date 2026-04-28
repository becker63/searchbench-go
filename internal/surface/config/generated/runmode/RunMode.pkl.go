// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package runmode

import (
	"encoding"
	"fmt"
)

type RunMode string

const (
	EvaluatorOnly       RunMode = "evaluator_only"
	WriterOptimization  RunMode = "writer_optimization"
	OptimizationKickoff RunMode = "optimization_kickoff"
)

// String returns the string representation of RunMode
func (rcv RunMode) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(RunMode)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for RunMode.
func (rcv *RunMode) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "evaluator_only":
		*rcv = EvaluatorOnly
	case "writer_optimization":
		*rcv = WriterOptimization
	case "optimization_kickoff":
		*rcv = OptimizationKickoff
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid RunMode`, str)
	}
	return nil
}
