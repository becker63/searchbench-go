package optimizer

import (
	"context"
	"os"

	optimizereino "github.com/becker63/searchbench-go/internal/agents/optimizer/eino"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Run executes one optimization manifest through the first optimizer loop.
func Run(ctx context.Context, request Request) (Record, error) {
	record := Record{ManifestPath: request.Resolve.ManifestPath}

	plan, err := Resolve(ctx, request.Resolve)
	if err != nil {
		return record, err
	}
	return RunResolved(ctx, plan, request)
}

// RunResolved executes one already-resolved optimization plan.
func RunResolved(ctx context.Context, plan Plan, request Request) (Record, error) {
	record := Record{ManifestPath: plan.ManifestPath}

	evidence, evidenceErr := loadEvidence(plan)
	if evidenceErr != nil {
		nextChallengerRecord := pureoptimizer.NextChallengerRecord{
			Failure: &pureoptimizer.Failure{
				Phase:   pureoptimizer.PhaseLoadParentEvidence,
				Kind:    pureoptimizer.FailureKindParentEvidenceFailed,
				Message: "load parent evidence",
				Cause:   evidenceErr,
			},
			Phases: []pureoptimizer.Phase{
				pureoptimizer.PhaseResolveOptimizerPlan,
				pureoptimizer.PhaseLoadParentEvidence,
			},
		}
		if bundlePath, bundleErr := writeBundle(ctx, plan, nextChallengerRecord); bundleErr == nil {
			record.BundlePath = bundlePath
		}
		record.Optimizer = nextChallengerRecord
		return record, nextChallengerRecord.Failure
	}

	spec := pureoptimizer.Spec{
		Target:   plan.Target,
		Agent:    plan.Agent,
		Evidence: evidence,
	}

	validator := request.ValidateProposal
	if validator == nil {
		validator = defaultValidateProposal
	}

	optimizerExecutor, err := optimizereino.NewOptimizer(optimizereino.OptimizerConfig{
		Model:            request.Model,
		RenderPrompt:     request.RenderPrompt,
		ValidateProposal: validator,
		RetryPolicy:      request.RetryPolicy,
	})
	if err != nil {
		nextChallengerRecord := pureoptimizer.NextChallengerRecord{
			Failure: &pureoptimizer.Failure{
				Phase:   pureoptimizer.PhasePrepareOptimizer,
				Kind:    pureoptimizer.FailureKindPrepareOptimizerFailed,
				Message: "construct optimizer executor",
				Cause:   err,
			},
			Phases: []pureoptimizer.Phase{
				pureoptimizer.PhaseResolveOptimizerPlan,
				pureoptimizer.PhaseLoadParentEvidence,
				pureoptimizer.PhasePrepareOptimizer,
			},
		}
		if bundlePath, bundleErr := writeBundle(ctx, plan, nextChallengerRecord); bundleErr == nil {
			record.BundlePath = bundlePath
		}
		record.Optimizer = nextChallengerRecord
		return record, nextChallengerRecord.Failure
	}

	nextChallengerRecord := optimizerExecutor.Run(ctx, spec)
	nextChallengerRecord.Phases = append(
		[]pureoptimizer.Phase{
			pureoptimizer.PhaseResolveOptimizerPlan,
			pureoptimizer.PhaseLoadParentEvidence,
			pureoptimizer.PhasePrepareOptimizer,
		},
		nextChallengerRecord.Phases...,
	)

	bundlePath, bundleErr := writeBundle(ctx, plan, nextChallengerRecord)
	record.BundlePath = bundlePath
	record.Optimizer = nextChallengerRecord
	if bundleErr != nil {
		return record, bundleErr
	}
	if nextChallengerRecord.Failure != nil {
		return record, nextChallengerRecord.Failure
	}
	return record, nil
}

func normalizeManifestPathError(manifestPath string, err error) error {
	if manifestPath == "" {
		return nil
	}
	if _, statErr := os.Stat(manifestPath); statErr != nil {
		return statErr
	}
	return nil
}
