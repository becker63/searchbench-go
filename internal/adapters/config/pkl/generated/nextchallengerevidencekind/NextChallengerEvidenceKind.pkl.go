// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package nextchallengerevidencekind

import (
	"encoding"
	"fmt"
)

// Evidence kinds the optimizer may pull from the parent bundle when `include` lists them.
// - `report_summary` — high-level report JSON / summary text.
// - `round_evidence` — structured round evidence document.
// - `objective_result` — prior objective scoring payload.
// - `challenger_policy` — the challenger policy source under evaluation.
type NextChallengerEvidenceKind string

const (
	ReportSummary    NextChallengerEvidenceKind = "report_summary"
	RoundEvidence    NextChallengerEvidenceKind = "round_evidence"
	ObjectiveResult  NextChallengerEvidenceKind = "objective_result"
	ChallengerPolicy NextChallengerEvidenceKind = "challenger_policy"
)

// String returns the string representation of NextChallengerEvidenceKind
func (rcv NextChallengerEvidenceKind) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(NextChallengerEvidenceKind)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for NextChallengerEvidenceKind.
func (rcv *NextChallengerEvidenceKind) UnmarshalBinary(data []byte) error {
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
		return fmt.Errorf(`illegal: "%s" is not a valid NextChallengerEvidenceKind`, str)
	}
	return nil
}
