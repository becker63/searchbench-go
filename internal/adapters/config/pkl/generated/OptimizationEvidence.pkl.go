// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

import (
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerdeniedevidencekind"
	"github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/optimizerevidencekind"
)

type OptimizationEvidence struct {
	From CompletedEvaluationBundleArtifact `pkl:"from"`

	Include []optimizerevidencekind.OptimizerEvidenceKind `pkl:"include"`

	Deny []optimizerdeniedevidencekind.OptimizerDeniedEvidenceKind `pkl:"deny"`
}
