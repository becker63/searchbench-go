package locale2e

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/app/evaluation"
	appOptimizer "github.com/becker63/searchbench-go/internal/app/optimizer"
)

// Run composes one local parent evaluation with one optimizer run.
func Run(ctx context.Context, request Request) (Result, error) {
	request = normalizeRequest(request)
	result := Result{}

	parentPlan, err := evaluation.Resolve(ctx, evaluation.ResolveRequest{
		ManifestPath:       request.EvaluationManifestPath,
		BundleRootOverride: parentEvaluationBundleRoot(request.BundleRootOverride),
		BundleID:           request.ParentEvaluationBundleID,
		Now:                request.Now,
	})
	if err != nil {
		return result, err
	}

	parentResult, err := evaluation.RunResolved(ctx, parentPlan, evaluation.Request{
		Resolve: evaluation.ResolveRequest{
			ManifestPath:       request.EvaluationManifestPath,
			BundleRootOverride: parentEvaluationBundleRoot(request.BundleRootOverride),
			BundleID:           request.ParentEvaluationBundleID,
			Now:                request.Now,
		},
		EvaluatorModelFactory: request.EvaluatorModelFactory,
		EvaluatorToolFactory:  request.EvaluatorToolFactory,
	})
	if err != nil {
		return result, err
	}
	result.ParentEvaluationBundle = string(parentResult.Bundle.Path)
	result.ParentEvaluationResult = &parentResult

	optimizerModelFactory := request.OptimizerModelFactory
	if optimizerModelFactory == nil {
		return result, fmt.Errorf("locale2e: optimizer model factory is required")
	}
	optimizerModel, err := optimizerModelFactory()
	if err != nil {
		return result, err
	}

	optimizerPlan, err := appOptimizer.Resolve(ctx, appOptimizer.ResolveRequest{
		ManifestPath:             request.OptimizationManifestPath,
		BundleRootOverride:       optimizerBundleRoot(request.BundleRootOverride),
		ParentBundlePathOverride: result.ParentEvaluationBundle,
		BundleID:                 request.OptimizerBundleID,
		Now:                      request.Now,
	})
	if err != nil {
		return result, err
	}

	optimizerResult, err := appOptimizer.RunResolved(ctx, optimizerPlan, appOptimizer.Request{
		Resolve: appOptimizer.ResolveRequest{
			ManifestPath:             request.OptimizationManifestPath,
			BundleRootOverride:       optimizerBundleRoot(request.BundleRootOverride),
			ParentBundlePathOverride: result.ParentEvaluationBundle,
			BundleID:                 request.OptimizerBundleID,
			Now:                      request.Now,
		},
		Model: optimizerModel,
	})
	result.OptimizerResult = &optimizerResult
	result.OptimizerBundle = optimizerResult.BundlePath
	if err != nil {
		return result, err
	}

	return result, nil
}
