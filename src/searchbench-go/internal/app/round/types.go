package round

import (
	"os"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"

	"github.com/becker63/searchbench-go/internal/pure/game"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	"github.com/becker63/searchbench-go/internal/pure/report"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
)

// OptimizerModelFactory constructs the optimizer model for one round.
type OptimizerModelFactory func() (model.ToolCallingChatModel, error)

// Input configures one round workflow.
//
// EvaluationManifestPath is the only required field.
type Input struct {
	EvaluationManifestPath string
	BundleRootOverride     string
	RoundID                string
	DisableRenderReport    bool
	Now                    func() time.Time

	EvaluatorModelFactory EvaluatorModelFactory
	EvaluatorToolFactory  EvaluatorToolFactory
	OptimizerModelFactory OptimizerModelFactory
	// OptimizerValidateProposal overrides default IC optimizer validation (full local pipeline).
	// Tests use a lightweight stub; production leaves this nil.
	OptimizerValidateProposal pureoptimizer.ValidateProposalFunc

	// DatasetMaterializeCacheDir triggers JetBrains LCA git materialization during
	// manifest resolution when non-empty.
	DatasetMaterializeCacheDir  string
	DatasetMaterializeRemoteURL string
}

// Resolved is the normalized round contract before match execution.
type Resolved struct {
	Game  game.Contract
	Round Plan
}

// MatchRecords captures the incumbent/challenger match execution output. It is
// the in-memory hand-off between EvaluateMatches and the downstream phase
// functions. It must not contain any bundle artifacts; persistence is reserved
// for WriteBundle.
type MatchRecords struct {
	Plan                Plan
	RoundReport         report.RoundReport
	EvaluatorExecutions []EvaluatorExecution
	MatchExecutions     []report.MatchExecutionRecord
}

// Record is the completed round workflow outcome.
type Record struct {
	Game  game.Contract
	Round pureround.Record

	RoundBundle string

	RoundResult *Result
}

func normalizeInput(input Input) Input {
	if input.Now == nil {
		input.Now = func() time.Time { return time.Now().UTC() }
	}
	if strings.TrimSpace(input.DatasetMaterializeCacheDir) == "" {
		input.DatasetMaterializeCacheDir = strings.TrimSpace(os.Getenv("SEARCHBENCH_MATERIALIZE_CACHE_DIR"))
	}
	if strings.TrimSpace(input.DatasetMaterializeRemoteURL) == "" {
		input.DatasetMaterializeRemoteURL = strings.TrimSpace(os.Getenv("SEARCHBENCH_MATERIALIZE_REMOTE_URL"))
	}
	return input
}

func roundBundleRoot(base string) string {
	if base == "" {
		return ""
	}
	return base
}
