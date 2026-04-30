package localrun

import (
	"path/filepath"

	artifact "github.com/becker63/searchbench-go/internal/adapters/artifact/fsbundle"
	appExperiment "github.com/becker63/searchbench-go/internal/app/experiment"
)

func bundleResolvedInput(plan appExperiment.ResolvedExperiment) artifact.ResolvedComparisonInput {
	return artifact.ResolvedComparisonInput{
		ManifestPath:   plan.ManifestPath,
		ExperimentName: plan.ExperimentName,
		Mode:           plan.Mode,
		Dataset: artifact.DatasetConfig{
			Kind:     plan.Dataset.Kind,
			Name:     plan.Dataset.Name,
			Config:   plan.Dataset.Config,
			Split:    plan.Dataset.Split,
			MaxItems: plan.Dataset.MaxItems,
		},
		Systems: plan.ComparePlan().ReportSpec().Systems,
		Tasks:   plan.Tasks,
		Parallelism: artifact.ParallelismConfig{
			Mode:       string(plan.Parallelism.Mode),
			MaxWorkers: plan.Parallelism.MaxWorkers,
			FailFast:   plan.Parallelism.FailFast,
		},
		Evaluator: artifact.EvaluatorConfig{
			Model: artifact.EvaluatorModelConfig{
				Provider:        plan.Evaluator.Model.Provider,
				Name:            plan.Evaluator.Model.Name,
				MaxOutputTokens: plan.Evaluator.Model.MaxOutputTokens,
			},
			Bounds: artifact.EvaluatorBoundsConfig{
				MaxModelTurns:  plan.Evaluator.Bounds.MaxModelTurns,
				MaxToolCalls:   plan.Evaluator.Bounds.MaxToolCalls,
				TimeoutSeconds: plan.Evaluator.Bounds.TimeoutSeconds,
			},
			Retry: artifact.RetryPolicyConfig{
				MaxAttempts:                plan.Evaluator.Retry.MaxAttempts,
				RetryOnModelError:          plan.Evaluator.Retry.RetryOnModelError,
				RetryOnToolFailure:         plan.Evaluator.Retry.RetryOnToolFailure,
				RetryOnFinalizationFailure: plan.Evaluator.Retry.RetryOnFinalizationFailure,
				RetryOnInvalidPrediction:   plan.Evaluator.Retry.RetryOnInvalidPrediction,
			},
		},
		Scoring: artifact.ScoringConfig{
			ObjectivePath: plan.Scoring.ObjectivePath,
			Evidence: artifact.EvidenceConfig{
				Current: plan.Scoring.CurrentEvidence,
				Parent:  cloneEvidenceRef(plan.Scoring.ParentEvidence),
			},
		},
		Output: artifact.OutputConfig{
			BundleRoot:        filepath.ToSlash(string(plan.Output.BundleCollectionPath)),
			BundleWriterRoot:  filepath.ToSlash(string(plan.Output.BundleWriterRoot)),
			ReportFormat:      plan.Output.ReportFormat,
			RenderHumanReport: plan.Output.RenderHumanReport,
			ResolvedPolicyPath: artifact.ResolvedPolicyPath{
				Baseline:  filepath.ToSlash(plan.Output.ResolvedPolicyPaths.Baseline),
				Candidate: filepath.ToSlash(plan.Output.ResolvedPolicyPaths.Candidate),
			},
		},
		ReportOptions: artifact.ReportOptions{
			Format: plan.Report.Format,
		},
	}
}
