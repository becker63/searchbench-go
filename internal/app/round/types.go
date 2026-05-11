package round

import (
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/components/model"

	appOptimizer "github.com/becker63/searchbench-go/internal/app/optimizer"
	"github.com/becker63/searchbench-go/internal/pure/game"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
)

// OptimizerModelFactory constructs the optimizer model for one round.
type OptimizerModelFactory func() (model.ToolCallingChatModel, error)

// Input configures one round workflow.
//
// EvaluationManifestPath is the only required field. When
// OptimizationManifestPath and OptimizerModelFactory are both omitted the
// round skips the optimizer and produces an evaluation-only bundle.
type Input struct {
	EvaluationManifestPath   string
	OptimizationManifestPath string
	BundleRootOverride       string
	RoundID                  string
	OptimizerBundleID        string
	DisableRenderReport      bool
	Now                      func() time.Time

	EvaluatorModelFactory EvaluatorModelFactory
	EvaluatorToolFactory  EvaluatorToolFactory
	OptimizerModelFactory OptimizerModelFactory
}

// Resolved is the normalized round contract before match execution.
type Resolved struct {
	Game  game.Contract
	Round Plan
}

// MatchRecords captures the incumbent/challenger match execution output.
type MatchRecords struct {
	Evaluation Result
}

// Record is the completed round workflow outcome.
type Record struct {
	Game  game.Contract
	Round pureround.Record

	RoundBundle     string
	OptimizerBundle string

	RoundResult          *Result
	NextChallengerResult *appOptimizer.Record
}

func normalizeInput(input Input) Input {
	if input.Now == nil {
		input.Now = func() time.Time { return time.Now().UTC() }
	}
	return input
}

func roundBundleRoot(base string) string {
	if base == "" {
		return ""
	}
	return base
}

func optimizerBundleRoot(base string) string {
	if base == "" {
		return ""
	}
	return filepath.Join(base, "optimizer")
}
