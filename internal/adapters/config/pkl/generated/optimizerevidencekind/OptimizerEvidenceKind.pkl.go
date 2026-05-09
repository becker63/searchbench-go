// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package optimizerevidencekind

import (
	"encoding"
	"fmt"
)

type OptimizerEvidenceKind string

const (
	ReportSummary    OptimizerEvidenceKind = "report_summary"
	RoundEvidence    OptimizerEvidenceKind = "round_evidence"
	ObjectiveResult  OptimizerEvidenceKind = "objective_result"
	ChallengerPolicy OptimizerEvidenceKind = "challenger_policy"
)

// String returns the string representation of OptimizerEvidenceKind
func (rcv OptimizerEvidenceKind) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(OptimizerEvidenceKind)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for OptimizerEvidenceKind.
func (rcv *OptimizerEvidenceKind) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "report_summary":
		*rcv = ReportSummary
	case "round_evidence":
		*rcv = RoundEvidence
	case "objective_result":
		*rcv = ObjectiveResult
	case "challenger_policy":
		*rcv = ChallengerPolicy
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid OptimizerEvidenceKind`, str)
	}
	return nil
}
