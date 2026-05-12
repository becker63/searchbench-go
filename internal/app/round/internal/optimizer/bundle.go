package optimizer

import (
	"context"
	"path/filepath"

	optimizebundle "github.com/becker63/searchbench-go/internal/agents/optimizer/bundle"
	optimizepolicy "github.com/becker63/searchbench-go/internal/agents/optimizer/policy"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// defaultValidateProposal delegates to the Python-policy validator bundled with the optimizer agent.
var defaultValidateProposal = optimizepolicy.Validate

// writeBundle delegates to optimizer-owned bundle persistence, projecting the in-memory
// Plan into manifest-derived ResolvedDocument metadata.
func writeBundle(ctx context.Context, plan Plan, result pureoptimizer.NextChallengerRecord) (string, error) {
	return optimizebundle.WriteBundle(ctx, optimizebundle.Request{
		BundleCollection: plan.BundleCollection,
		BundleID:         plan.BundleID,
		CreatedAt:        plan.CreatedAt,
		ParentBundle:     string(plan.ParentBundle.BundlePath),
		OutputArtifact:   plan.Target.OutputName,
		Resolved: optimizebundle.ResolvedDocument{
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
