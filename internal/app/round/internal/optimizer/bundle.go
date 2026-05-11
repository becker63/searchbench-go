package optimizer

import (
	"context"
	"path/filepath"

	optimizerfs "github.com/becker63/searchbench-go/internal/adapters/optimizer/fs"
	pythonpolicy "github.com/becker63/searchbench-go/internal/adapters/policy/python"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// defaultValidateProposal delegates to the python policy adapter.
var defaultValidateProposal = pythonpolicy.Validate

// writeBundle delegates to the optimizer/fs adapter, projecting the in-memory
// Plan into the adapter's manifest-derived ResolvedDocument so the adapter
// holds the only fs-side ownership.
func writeBundle(ctx context.Context, plan Plan, result pureoptimizer.NextChallengerRecord) (string, error) {
	return optimizerfs.WriteBundle(ctx, optimizerfs.Request{
		BundleCollection: plan.BundleCollection,
		BundleID:         plan.BundleID,
		CreatedAt:        plan.CreatedAt,
		ParentBundle:     string(plan.ParentBundle.BundlePath),
		OutputArtifact:   plan.Target.OutputName,
		Resolved: optimizerfs.ResolvedDocument{
			ManifestPath:     plan.ManifestPath,
			RoundName:        plan.RoundName,
			Mode:             "optimization",
			ParentRound:      plan.ParentBundle,
			Target:           plan.Target,
			Agent:            plan.Agent,
			IncludedEvidence: plan.IncludedEvidence,
			DeniedEvidence:   plan.DeniedEvidence,
			InputPolicyPath:  filepath.ToSlash(plan.InputPolicy.Path),
			OutputBundlePath: filepath.ToSlash(plan.ExpectedBundlePath),
		},
		Result: result,
	})
}
