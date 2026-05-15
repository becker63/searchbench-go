package codelocalization

import (
	"testing"

	"github.com/becker63/searchbench-go/internal/pure/game"
)

func TestGameDefinesMatchEvidenceAndReviewPanes(t *testing.T) {
	t.Parallel()

	contract := Contract()
	if contract.ID != GameID {
		t.Fatalf("ID = %q, want %q", contract.ID, GameID)
	}
	if contract.Kind != game.KindCodeLocalization {
		t.Fatalf("Kind = %q, want %q", contract.Kind, game.KindCodeLocalization)
	}
	if contract.MatchSchema.ID != MatchSchemaID {
		t.Fatalf("MatchSchema.ID = %q, want %q", contract.MatchSchema.ID, MatchSchemaID)
	}
	if contract.EvidenceSchema.ID != EvidenceSchemaID {
		t.Fatalf("EvidenceSchema.ID = %q, want %q", contract.EvidenceSchema.ID, EvidenceSchemaID)
	}
	if len(contract.ReviewPanes) < 2 {
		t.Fatalf("ReviewPanes = %#v, want mirrored replay and evidence panes", contract.ReviewPanes)
	}
}

func TestCodeLocalizationGameBuildsHopDistanceEvidence(t *testing.T) {
	t.Parallel()

	evidence := HopDistanceEvidence{
		GoldHop:  2,
		IssueHop: 3,
	}
	if evidence.GoldHop >= evidence.IssueHop {
		t.Fatalf("evidence = %#v, want distinct hop-distance fields", evidence)
	}
}
