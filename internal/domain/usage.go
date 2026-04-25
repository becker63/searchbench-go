package domain

// TokenCount is a normalized token count.
//
// Keep it separate from int so token/cost code is explicit.
type TokenCount int64

// UsageSummary is normalized run-level usage.
//
// Provider-specific usage details can exist elsewhere. This type is the stable
// Searchbench shape used by scoring and reports.
type UsageSummary struct {
	InputTokens  TokenCount `json:"input_tokens"`
	OutputTokens TokenCount `json:"output_tokens"`
	TotalTokens  TokenCount `json:"total_tokens"`
	CostUSD      float64    `json:"cost_usd,omitempty"`
}

// Empty reports whether no usage was recorded.
func (u UsageSummary) Empty() bool {
	return u.InputTokens == 0 &&
		u.OutputTokens == 0 &&
		u.TotalTokens == 0 &&
		u.CostUSD == 0
}
