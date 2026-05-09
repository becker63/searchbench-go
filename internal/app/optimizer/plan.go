package optimizer

import (
	"time"

	"github.com/cloudwego/eino/components/model"

	executoreino "github.com/becker63/searchbench-go/internal/adapters/executor/eino"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// Request configures one optimizer run.
type Request struct {
	ManifestPath       string
	BundleRootOverride string
	BundleID           string
	Now                func() time.Time
	Model              model.ToolCallingChatModel
	ValidateProposal   executoreino.ValidateProposalFunc
	RenderPrompt       executoreino.RenderOptimizerPromptFunc
	RetryPolicy        *pureoptimizer.RetryPolicy
}

// Result is the app-level optimizer run outcome.
type Result struct {
	ManifestPath string
	BundlePath   string
	Optimizer    pureoptimizer.Result
}

// Plan is the resolved optimization plan before evidence loading.
type Plan struct {
	ManifestPath       string
	ExperimentName     string
	CreatedAt          time.Time
	BundleID           string
	BundleCollection   string
	BundleWriterRoot   string
	ExpectedBundlePath string
	Agent              pureoptimizer.AgentConfig
	Target             pureoptimizer.Target
	ParentBundle       pureoptimizer.ParentRunRef
	InputPolicy        InputPolicyPlan
	IncludedEvidence   []string
	DeniedEvidence     []string
}

// InputPolicyPlan is the unresolved input policy artifact reference.
type InputPolicyPlan struct {
	ArtifactID  string
	Path        string
	InterfaceID string
}
