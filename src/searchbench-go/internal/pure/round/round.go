// Package round defines pure round records and lineage.
package round

import (
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/game"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// ID identifies one immutable round under a game.
type ID string

// BundleRef identifies a completed durable round bundle.
type BundleRef struct {
	RoundID ID              `json:"round_id"`
	Path    domain.HostPath `json:"path"`
}

// Lineage records parent-round evidence for cross-round comparison.
type Lineage struct {
	ParentRound *BundleRef `json:"parent_round,omitempty"`
}

// Record is the completed pure round outcome.
type Record struct {
	GameID          game.ID                     `json:"game_id"`
	RoundID         string                      `json:"round_id"`
	BundlePath      domain.HostPath             `json:"bundle_path"`
	Evidence        score.RoundEvidenceDocument `json:"evidence"`
	ObjectiveResult *score.ObjectiveResult      `json:"objective_result,omitempty"`
	Decision        report.Decision             `json:"decision"`
	NextChallenger  bool                        `json:"next_challenger"`
}
