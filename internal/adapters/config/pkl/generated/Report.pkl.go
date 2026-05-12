// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/reportformat"

// Human- and machine-readable report outputs for the round bundle.
type Report struct {
	// Which report artifacts to materialize (JSON and/or text summaries).
	Formats []reportformat.ReportFormat `pkl:"formats"`
}
