package round

import (
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	evaluatoreino "github.com/becker63/searchbench-go/internal/agents/evaluator/eino"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
	"github.com/becker63/searchbench-go/internal/pure/usage"
)

// EvaluatorModelFactory constructs one evaluator model for one run spec.
type EvaluatorModelFactory func(spec run.Spec) (model.ToolCallingChatModel, error)

// EvaluatorToolFactory constructs the evaluator tool set for one run spec.
type EvaluatorToolFactory func(spec run.Spec) ([]tool.BaseTool, error)

// evaluationRequest configures one manifest-driven local evaluation run.
type evaluationRequest struct {
	Resolve evaluationResolveRequest

	PklCommand          []string
	DisableRenderReport bool

	EvaluatorModelFactory EvaluatorModelFactory
	EvaluatorToolFactory  EvaluatorToolFactory

	HashRegistry *usage.HashRegistry
}

// EvaluatorExecution is one recorded evaluator-backed run inside the local
// comparison flow.
type EvaluatorExecution struct {
	Role    domain.Role
	MatchID domain.MatchID
	RunID   domain.RunID
	Result  evaluatoreino.Result
}

// Result is the completed manifest-driven local evaluation run.
type Result struct {
	ManifestPath        string
	Bundle              bundlefs.RoundBundleRef
	ReportID            domain.ReportID
	RoundReport         report.RoundReport
	RoundEvidence       score.RoundEvidenceDocument
	ObjectiveResult     *score.ObjectiveResult
	EvaluatorExecutions []EvaluatorExecution
	MatchExecutions     []report.MatchExecutionRecord
	HashRegistry        *usage.HashRegistry
}
