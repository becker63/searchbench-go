// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import (
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/nextchallengerevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerdeniedevidencekind"
)

// Evidence inclusion and denial rules for optimizer prompts.
type NextChallengerEvidence struct {
	// Parent bundle artifact that evidence is loaded from; must match optimization parent wiring (validated in Go).
	From CompletedRoundBundleArtifact `pkl:"from"`

	// Evidence kinds pulled into the prompt when present in the parent bundle.
	Include []nextchallengerevidencekind.NextChallengerEvidenceKind `pkl:"include"`

	// Evidence channels that must never be inlined into the optimizer prompt.
	Deny []optimizerdeniedevidencekind.OptimizerDeniedEvidenceKind `pkl:"deny"`
}
