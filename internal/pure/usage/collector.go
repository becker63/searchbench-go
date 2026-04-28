package usage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// ReportedUsage is provider-reported token metadata normalized into the
// harness-owned collector input shape.
type ReportedUsage struct {
	InputTokens  MaybeTokenCount
	OutputTokens MaybeTokenCount
	TotalTokens  MaybeTokenCount
}

// StartEvent is the provider-neutral model-call start payload.
type StartEvent struct {
	Phase    string
	Node     string
	Provider string
	Model    string
	Input    []string
}

// EndEvent is the provider-neutral model-call end payload.
type EndEvent struct {
	Provider string
	Model    string
	Output   []string
	Reported ReportedUsage
}

// Config defines how a collector should normalize usage for one evaluator run.
type Config struct {
	Tokenizer       Tokenizer
	DefaultProvider string
	DefaultModel    string
}

type pendingCall struct {
	phase    string
	node     string
	provider string
	model    string
	input    []string
}

// Collector accumulates per-model-call usage facts and normalizes them into
// canonical harness-owned records.
type Collector struct {
	tokenizer       Tokenizer
	defaultProvider string
	defaultModel    string

	mu      sync.Mutex
	nextID  int
	pending map[int]pendingCall
	records []Record
}

// NewCollector constructs a usage collector for one evaluator run.
func NewCollector(config Config) (*Collector, error) {
	tokenizer := config.Tokenizer
	if tokenizer == nil {
		tokenizer = WhitespaceTokenizer{}
	}

	return &Collector{
		tokenizer:       tokenizer,
		defaultProvider: config.DefaultProvider,
		defaultModel:    config.DefaultModel,
		pending:         make(map[int]pendingCall),
	}, nil
}

// StartCall records the model input for one model call and returns an opaque
// call identifier for later completion.
func (c *Collector) StartCall(event StartEvent) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++
	c.pending[id] = pendingCall{
		phase:    event.Phase,
		node:     event.Node,
		provider: choose(event.Provider, c.defaultProvider),
		model:    choose(event.Model, c.defaultModel),
		input:    append([]string(nil), event.Input...),
	}
	return id
}

// EndCall finalizes one model-call usage record.
func (c *Collector) EndCall(id int, event EndEvent) {
	c.mu.Lock()
	pending, ok := c.pending[id]
	if ok {
		delete(c.pending, id)
	}
	c.mu.Unlock()

	if !ok {
		return
	}

	record := Record{
		Phase:    pending.phase,
		Node:     pending.node,
		Provider: choose(event.Provider, pending.provider),
		Model:    choose(event.Model, pending.model),
	}

	record.ReportedInputTokens = event.Reported.InputTokens
	record.ReportedOutputTokens = event.Reported.OutputTokens
	record.ReportedTotalTokens = event.Reported.TotalTokens

	var inputErr error
	if !event.Reported.InputTokens.Set {
		var inputEstimate domain.TokenCount
		inputEstimate, inputErr = c.tokenizer.CountStrings(pending.input)
		if inputErr == nil {
			record.EstimatedInputTokens = MaybeTokenCount{Value: inputEstimate, Set: true}
		}
	}

	var outputErr error
	if !event.Reported.OutputTokens.Set {
		var outputEstimate domain.TokenCount
		outputEstimate, outputErr = c.tokenizer.CountStrings(event.Output)
		if outputErr == nil {
			record.EstimatedOutputTokens = MaybeTokenCount{Value: outputEstimate, Set: true}
		}
	}

	record.InputTokens = chooseCount(event.Reported.InputTokens, record.EstimatedInputTokens)
	record.OutputTokens = chooseCount(event.Reported.OutputTokens, record.EstimatedOutputTokens)

	switch {
	case event.Reported.TotalTokens.Set:
		record.TotalTokens = event.Reported.TotalTokens.Value
	case record.InputTokens > 0 || record.OutputTokens > 0:
		record.TotalTokens = record.InputTokens + record.OutputTokens
	}

	if inputErr != nil {
		record.Issues = append(record.Issues, issueForEstimateError(inputErr, "estimate input tokens"))
	}
	if outputErr != nil {
		record.Issues = append(record.Issues, issueForEstimateError(outputErr, "estimate output tokens"))
	}

	if !hasCanonicalCounts(record) || !hasCompleteRecord(record) {
		record.Issues = append(record.Issues, Issue{
			Kind:    IssueIncompleteUsage,
			Message: "usage record is incomplete",
		})
	}

	record.Source = classifySource(record)

	c.mu.Lock()
	c.records = append(c.records, record)
	c.mu.Unlock()
}

// Records returns a stable snapshot of all completed usage records.
func (c *Collector) Records() []Record {
	c.mu.Lock()
	defer c.mu.Unlock()

	records := make([]Record, len(c.records))
	copy(records, c.records)
	for i := range records {
		records[i].Issues = append([]Issue(nil), records[i].Issues...)
	}
	return records
}

// Summary returns the aggregated run-level usage summary.
func (c *Collector) Summary() Summary {
	records := c.Records()
	summary := Summary{
		CallCount: len(records),
	}
	for _, record := range records {
		summary.InputTokens += record.InputTokens
		summary.OutputTokens += record.OutputTokens
		summary.TotalTokens += record.TotalTokens
		if !record.Complete() {
			summary.IncompleteRecords++
			summary.Issues = append(summary.Issues, record.Issues...)
		}
	}
	return summary
}

func choose(primary string, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func chooseCount(primary MaybeTokenCount, fallback MaybeTokenCount) domain.TokenCount {
	if primary.Set {
		return primary.Value
	}
	if fallback.Set {
		return fallback.Value
	}
	return 0
}

func classifySource(record Record) Source {
	reported := record.ReportedInputTokens.Set || record.ReportedOutputTokens.Set || record.ReportedTotalTokens.Set
	estimated := record.EstimatedInputTokens.Set || record.EstimatedOutputTokens.Set

	switch {
	case reported && estimated:
		return SourceMixed
	case reported:
		return SourceReported
	case estimated:
		return SourceEstimated
	default:
		return SourceUnavailable
	}
}

func issueForEstimateError(err error, action string) Issue {
	if err == nil {
		return Issue{}
	}
	if errors.Is(err, ErrTokenizerUnavailable) {
		return Issue{
			Kind:    IssueTokenizerUnavailable,
			Message: fmt.Sprintf("%s: %v", action, err),
		}
	}
	return Issue{
		Kind:    IssueEstimationFailed,
		Message: fmt.Sprintf("%s: %v", action, err),
	}
}

func hasCanonicalCounts(record Record) bool {
	return record.InputTokens > 0 || record.OutputTokens > 0 || record.TotalTokens > 0
}

func hasCompleteRecord(record Record) bool {
	return (record.InputTokens > 0 || record.ReportedInputTokens.Set || record.EstimatedInputTokens.Set) &&
		(record.OutputTokens > 0 || record.ReportedOutputTokens.Set || record.EstimatedOutputTokens.Set) &&
		(record.TotalTokens > 0 || record.ReportedTotalTokens.Set || (record.InputTokens > 0 && record.OutputTokens > 0))
}
