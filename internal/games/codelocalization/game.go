// Package codelocalization contains pure concepts for the code-localization game.
package codelocalization

import "github.com/becker63/searchbench-go/internal/pure/game"

const (
	GameID           game.ID = "code-localization"
	MatchSchemaID            = "searchbench.games.code_localization.match.v1"
	EvidenceSchemaID         = "searchbench.games.code_localization.evidence.v1"
)

// EvidenceName identifies one code-localization evidence field.
type EvidenceName string

const (
	EvidenceHopDistance   EvidenceName = "hop_distance"
	EvidenceTargetFiles   EvidenceName = "target_files"
	EvidenceTargetSymbols EvidenceName = "target_symbols"
)

// HopDistanceEvidence captures how far a policy landed from the expected code
// localization target.
type HopDistanceEvidence struct {
	GoldHop  float64 `json:"gold_hop"`
	IssueHop float64 `json:"issue_hop"`
}

// TargetEvidence describes code targets surfaced during review.
type TargetEvidence struct {
	Files   []string `json:"files,omitempty"`
	Symbols []string `json:"symbols,omitempty"`
}

// Contract returns the pure contract for the first SearchBench game.
func Contract() game.Contract {
	return game.Contract{
		ID:   GameID,
		Kind: game.KindCodeLocalization,
		Name: "Code Localization",
		MatchSchema: game.SchemaRef{
			ID:      MatchSchemaID,
			Version: "v1",
		},
		EvidenceSchema: game.SchemaRef{
			ID:      EvidenceSchemaID,
			Version: "v1",
		},
		ReviewPanes: []game.ReviewPane{
			{
				ID:          "mirrored_replay",
				Name:        "Mirrored Replay",
				Description: "Compare incumbent and challenger graph traversal on the same localization match.",
			},
			{
				ID:          "evidence",
				Name:        "Evidence",
				Description: "Review hop distance, target coverage, usage, and protected regressions.",
			},
		},
		MatchOutcomeSemantics: game.MatchOutcomeSemantics{
			WinLabel:  "localized target",
			LossLabel: "missed target",
		},
	}
}
