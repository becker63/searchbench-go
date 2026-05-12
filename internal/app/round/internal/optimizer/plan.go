package optimizer

import (
	"time"

	"github.com/cloudwego/eino/components/model"

	optimizereino "github.com/becker63/searchbench-go/internal/agents/optimizer/eino"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

// ResolveRequest configures one optimizer plan resolution.
type ResolveRequest struct {
	ManifestPath             string
	BundleRootOverride       string
	ParentBundlePathOverride string
	BundleID                 string
	Now                      func() time.Time
}

// Request configures one next-challenger run.
type Request struct {
	Resolve ResolveRequest

	Model            model.ToolCallingChatModel
	ValidateProposal optimizereino.ValidateProposalFunc
	RenderPrompt     optimizereino.RenderOptimizerPromptFunc
	RetryPolicy      *pureoptimizer.RetryPolicy
}

// Record is the app-level next-challenger run outcome.
type Record struct {
	ManifestPath string
	BundlePath   string
	Optimizer    pureoptimizer.NextChallengerRecord
}

// Plan is the resolved optimization plan before evidence loading.
type Plan struct {
	ManifestPath       string
	RoundName          string
	CreatedAt          time.Time
	BundleID           string
	BundleCollection   string
	BundleWriterRoot   string
	ExpectedBundlePath string
	Agent              pureoptimizer.AgentConfig
	Target             pureoptimizer.NextChallengerTarget
	ParentBundle       pureoptimizer.ParentRoundRef
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
