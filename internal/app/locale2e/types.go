package locale2e

import (
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/components/model"

	"github.com/becker63/searchbench-go/internal/app/evaluation"
	appOptimizer "github.com/becker63/searchbench-go/internal/app/optimizer"
)

// OptimizerModelFactory constructs the optimizer model for one local e2e run.
type OptimizerModelFactory func() (model.ToolCallingChatModel, error)

// Request configures one local evaluation -> optimization workflow.
type Request struct {
	EvaluationManifestPath   string
	OptimizationManifestPath string
	BundleRootOverride       string
	ParentEvaluationBundleID string
	OptimizerBundleID        string
	Now                      func() time.Time

	EvaluatorModelFactory evaluation.EvaluatorModelFactory
	EvaluatorToolFactory  evaluation.EvaluatorToolFactory
	OptimizerModelFactory OptimizerModelFactory
}

// Result is the composed local workflow outcome.
type Result struct {
	ParentEvaluationBundle string
	OptimizerBundle        string

	ParentEvaluationResult *evaluation.Result
	OptimizerResult        *appOptimizer.Result
}

func normalizeRequest(request Request) Request {
	if request.Now == nil {
		request.Now = func() time.Time { return time.Now().UTC() }
	}
	return request
}

func parentEvaluationBundleRoot(base string) string {
	if base == "" {
		return ""
	}
	return filepath.Join(base, "parent-evaluation")
}

func optimizerBundleRoot(base string) string {
	if base == "" {
		return ""
	}
	return filepath.Join(base, "optimizer")
}
