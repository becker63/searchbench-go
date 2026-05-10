// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import (
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/nextchallengerevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerdeniedevidencekind"
)

type NextChallengerEvidence struct {
	From CompletedRoundBundleArtifact `pkl:"from"`

	Include []nextchallengerevidencekind.NextChallengerEvidenceKind `pkl:"include"`

	Deny []optimizerdeniedevidencekind.OptimizerDeniedEvidenceKind `pkl:"deny"`
}
