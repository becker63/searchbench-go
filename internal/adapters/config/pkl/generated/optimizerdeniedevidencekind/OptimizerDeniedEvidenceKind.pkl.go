// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package optimizerdeniedevidencekind

import (
	"encoding"
	"fmt"
)

// Labels for evidence the optimizer prompt must never include (leakage / safety).
// `gold_labels` — canonical answers or gold labels from the dataset.
// `oracle_files` — file lists or paths from oracle-only fields.
// `raw_dataset_answers` — raw supervision fields from dataset rows.
type OptimizerDeniedEvidenceKind string

const (
	GoldLabels        OptimizerDeniedEvidenceKind = "gold_labels"
	OracleFiles       OptimizerDeniedEvidenceKind = "oracle_files"
	RawDatasetAnswers OptimizerDeniedEvidenceKind = "raw_dataset_answers"
)

// String returns the string representation of OptimizerDeniedEvidenceKind
func (rcv OptimizerDeniedEvidenceKind) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(OptimizerDeniedEvidenceKind)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for OptimizerDeniedEvidenceKind.
func (rcv *OptimizerDeniedEvidenceKind) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "gold_labels":
		*rcv = GoldLabels
	case "oracle_files":
		*rcv = OracleFiles
	case "raw_dataset_answers":
		*rcv = RawDatasetAnswers
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid OptimizerDeniedEvidenceKind`, str)
	}
	return nil
}
