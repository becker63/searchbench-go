// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package optimizerevidencekind

import (
	"encoding"
	"fmt"
)

type OptimizerEvidenceKind string

const (
	ReportSummary   OptimizerEvidenceKind = "report_summary"
	ScoreEvidence   OptimizerEvidenceKind = "score_evidence"
	ObjectiveResult OptimizerEvidenceKind = "objective_result"
	CandidatePolicy OptimizerEvidenceKind = "candidate_policy"
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
	case "score_evidence":
		*rcv = ScoreEvidence
	case "objective_result":
		*rcv = ObjectiveResult
	case "candidate_policy":
		*rcv = CandidatePolicy
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid OptimizerEvidenceKind`, str)
	}
	return nil
}
