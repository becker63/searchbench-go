package execution

import "github.com/becker63/searchbench-go/internal/pure/domain"

// Spec is a planned request to run one system on one match.
//
// Conceptually:
//
//	Spec = MatchSpec + SystemSpec
//
// It should be deterministic and serializable. Runtime state belongs in
// PreparedRun or ExecutedRun, not here.
type Spec struct {
	ID                domain.RunID         `json:"id"`
	Match             domain.MatchSpec     `json:"match"`
	System            domain.SystemSpec    `json:"system"`
	EvaluatorAppendix EvaluatorRunAppendix `json:"evaluator_appendix,omitempty"`
}

// EvaluatorRunAppendix carries manifest evaluator text that must reach the
// evaluator prompt without altering SystemSpec (which is per policy role).
type EvaluatorRunAppendix struct {
	// SystemPrompt is optional manifest-controlled text (trimmed); escaped when rendered.
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// NewSpec constructs a planned run request from one match and one executable
// system.
func NewSpec(id domain.RunID, match domain.MatchSpec, system domain.SystemSpec) Spec {
	return Spec{
		ID:     id,
		Match:  match,
		System: system,
	}
}
