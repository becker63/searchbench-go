// Package game defines the pure SearchBench game contract.
package game

// ID identifies one SearchBench game.
type ID string

// Kind identifies the domain family for a game.
type Kind string

const (
	// KindCodeLocalization is the first concrete SearchBench game.
	KindCodeLocalization Kind = "code_localization"
)

// SchemaRef names a versioned game-owned schema.
type SchemaRef struct {
	ID      string `json:"id"`
	Version string `json:"version,omitempty"`
}

// ReviewPane declares one game-specific review surface.
type ReviewPane struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// MatchOutcomeSemantics describes how a game interprets one match outcome.
type MatchOutcomeSemantics struct {
	WinLabel  string `json:"win_label"`
	LossLabel string `json:"loss_label"`
}

// Contract is the minimal pure contract needed to run and review a round.
type Contract struct {
	ID                    ID                    `json:"id"`
	Kind                  Kind                  `json:"kind"`
	Name                  string                `json:"name"`
	MatchSchema           SchemaRef             `json:"match_schema"`
	EvidenceSchema        SchemaRef             `json:"evidence_schema"`
	ReviewPanes           []ReviewPane          `json:"review_panes,omitempty"`
	MatchOutcomeSemantics MatchOutcomeSemantics `json:"match_outcome_semantics"`
}
