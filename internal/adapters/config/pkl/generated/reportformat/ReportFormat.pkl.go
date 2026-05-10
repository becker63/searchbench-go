// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package reportformat

import (
	"encoding"
	"fmt"
)

type ReportFormat string

const (
	Json ReportFormat = "json"
	Text ReportFormat = "text"
)

// String returns the string representation of ReportFormat
func (rcv ReportFormat) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(ReportFormat)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for ReportFormat.
func (rcv *ReportFormat) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "json":
		*rcv = Json
	case "text":
		*rcv = Text
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid ReportFormat`, str)
	}
	return nil
}
