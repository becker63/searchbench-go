package optimizer

import (
	"context"
	"errors"
	"fmt"
	"os"

	executoreino "github.com/becker63/searchbench-go/internal/adapters/executor/eino"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Run executes one optimization manifest through the first optimizer loop.
func Run(ctx context.Context, request Request) (Result, error) {
	result := Result{ManifestPath: request.ManifestPath}

	plan, err := Resolve(ctx, request)
	if err != nil {
		return result, err
	}

	evidence, evidenceErr := loadEvidence(plan)
	if evidenceErr != nil {
		optimizerResult := pureoptimizer.Result{
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
		if bundlePath, bundleErr := writeBundle(ctx, plan, optimizerResult); bundleErr == nil {
			result.BundlePath = bundlePath
		}
		result.Optimizer = optimizerResult
		return result, optimizerResult.Failure
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

	optimizerExecutor, err := executoreino.NewOptimizer(executoreino.OptimizerConfig{
		Model:            request.Model,
		RenderPrompt:     request.RenderPrompt,
		ValidateProposal: validator,
		RetryPolicy:      request.RetryPolicy,
	})
	if err != nil {
		optimizerResult := pureoptimizer.Result{
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
		if bundlePath, bundleErr := writeBundle(ctx, plan, optimizerResult); bundleErr == nil {
			result.BundlePath = bundlePath
		}
		result.Optimizer = optimizerResult
		return result, optimizerResult.Failure
	}

	optimizerResult := optimizerExecutor.Run(ctx, spec)
	optimizerResult.Phases = append(
		[]pureoptimizer.Phase{
			pureoptimizer.PhaseResolveOptimizerPlan,
			pureoptimizer.PhaseLoadParentEvidence,
			pureoptimizer.PhasePrepareOptimizer,
		},
		optimizerResult.Phases...,
	)

	bundlePath, bundleErr := writeBundle(ctx, plan, optimizerResult)
	result.BundlePath = bundlePath
	result.Optimizer = optimizerResult
	if bundleErr != nil {
		return result, bundleErr
	}
	if optimizerResult.Failure != nil {
		return result, optimizerResult.Failure
	}
	return result, nil
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

func wrapRunError(request Request, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, os.ErrNotExist) {
		if absErr := normalizeManifestPathError(request.ManifestPath, err); absErr != nil {
			return absErr
		}
	}
	return fmt.Errorf("optimizer run: %w", err)
}
