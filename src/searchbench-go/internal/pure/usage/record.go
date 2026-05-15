package usage

import "github.com/becker63/searchbench-go/internal/pure/domain"

// Source identifies where the canonical token counts in a usage record came
// from.
type Source string

const (
	SourceReported    Source = "reported"
	SourceEstimated   Source = "estimated"
	SourceMixed       Source = "mixed"
	SourceUnavailable Source = "unavailable"
)

// IssueKind identifies a non-fatal usage accounting problem.
type IssueKind string

const (
	IssueTokenizerUnavailable IssueKind = "tokenizer_unavailable"
	IssueEstimationFailed     IssueKind = "estimation_failed"
	IssueIncompleteUsage      IssueKind = "incomplete_usage"
)

// Issue records one usage accounting problem without invalidating evaluator
// success.
type Issue struct {
	Kind    IssueKind `json:"kind"`
	Message string    `json:"message,omitempty"`
}

// MaybeTokenCount is an optional normalized token count.
type MaybeTokenCount struct {
	Value domain.TokenCount `json:"value"`
	Set   bool              `json:"set"`
}

// ReportedCount converts a raw provider-reported integer count into the
// harness-owned optional token-count shape.
func ReportedCount(value int64) MaybeTokenCount {
	return MaybeTokenCount{
		Value: domain.TokenCount(value),
		Set:   true,
	}
}

// Record is the canonical per-model-call usage record owned by the harness.
//
// It deliberately avoids provider-specific blobs and Eino callback types.
type Record struct {
	Phase string `json:"phase,omitempty"`
	Node  string `json:"node,omitempty"`

	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`

	Source Source `json:"source"`

	InputTokens  domain.TokenCount `json:"input_tokens"`
	OutputTokens domain.TokenCount `json:"output_tokens"`
	TotalTokens  domain.TokenCount `json:"total_tokens"`

	ReportedInputTokens  MaybeTokenCount `json:"reported_input_tokens,omitempty"`
	ReportedOutputTokens MaybeTokenCount `json:"reported_output_tokens,omitempty"`
	ReportedTotalTokens  MaybeTokenCount `json:"reported_total_tokens,omitempty"`

	EstimatedInputTokens  MaybeTokenCount `json:"estimated_input_tokens,omitempty"`
	EstimatedOutputTokens MaybeTokenCount `json:"estimated_output_tokens,omitempty"`

	Issues []Issue `json:"issues,omitempty"`
}

// Complete reports whether the usage record has no recorded accounting issues.
func (r Record) Complete() bool {
	return len(r.Issues) == 0 && r.Source != SourceUnavailable
}
