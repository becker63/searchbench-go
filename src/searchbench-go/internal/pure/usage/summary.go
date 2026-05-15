package usage

import "github.com/becker63/searchbench-go/internal/pure/domain"

// Summary is the aggregated evaluator-run usage view built from per-call
// records.
type Summary struct {
	CallCount         int               `json:"call_count"`
	InputTokens       domain.TokenCount `json:"input_tokens"`
	OutputTokens      domain.TokenCount `json:"output_tokens"`
	TotalTokens       domain.TokenCount `json:"total_tokens"`
	IncompleteRecords int               `json:"incomplete_records"`
	Issues            []Issue           `json:"issues,omitempty"`
}

// Complete reports whether every recorded model call had complete usage data.
func (s Summary) Complete() bool {
	return s.IncompleteRecords == 0 && len(s.Issues) == 0
}

// DomainSummary converts the harness-owned usage summary into the stable
// report/scoring shape already used by executed runs.
func (s Summary) DomainSummary() domain.UsageSummary {
	return domain.UsageSummary{
		InputTokens:  s.InputTokens,
		OutputTokens: s.OutputTokens,
		TotalTokens:  s.TotalTokens,
	}
}
