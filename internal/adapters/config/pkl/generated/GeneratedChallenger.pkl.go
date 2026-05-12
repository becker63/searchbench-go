// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import (
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/nextchallengerevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerdeniedevidencekind"
)

// Request to materialize the challenger through the optimizer seam before the
// round is evaluated.
type GeneratedChallenger struct {
	Optimizer Optimizer `pkl:"optimizer"`

	ArtifactName string `pkl:"artifactName"`

	Include []nextchallengerevidencekind.NextChallengerEvidenceKind `pkl:"include"`

	Deny []optimizerdeniedevidencekind.OptimizerDeniedEvidenceKind `pkl:"deny"`
}
