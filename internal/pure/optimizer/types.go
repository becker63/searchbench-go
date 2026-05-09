package optimizer

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Phase names one stable optimizer lifecycle phase.
type Phase string

const (
	PhaseResolveOptimizerPlan  Phase = "resolve_optimizer_plan"
	PhaseLoadParentEvidence    Phase = "load_parent_evidence"
	PhasePrepareOptimizer      Phase = "prepare_optimizer"
	PhaseRenderOptimizerPrompt Phase = "render_optimizer_prompt"
	PhaseRunOptimizer          Phase = "run_optimizer"
	PhaseFinalizeProposal      Phase = "finalize_policy_proposal"
	PhaseWriteCandidateStage   Phase = "write_candidate_stage"
	PhaseRunPolicyPipeline     Phase = "run_policy_pipeline"
	PhaseClassifyPipeline      Phase = "classify_policy_pipeline"
	PhasePrepareRetry          Phase = "prepare_retry"
	PhaseAcceptProposal        Phase = "accept_proposal"
	PhaseWriteOptimizerBundle  Phase = "write_optimizer_bundle"
	PhaseComplete              Phase = "complete"
	PhaseExhausted             Phase = "exhausted"
)

// FailureKind is one stable optimizer failure category.
type FailureKind string

const (
	FailureKindResolvePlanFailed            FailureKind = "resolve_optimizer_plan_failed"
	FailureKindParentEvidenceFailed         FailureKind = "parent_evidence_failed"
	FailureKindPrepareOptimizerFailed       FailureKind = "prepare_optimizer_failed"
	FailureKindOptimizerPromptFailed        FailureKind = "optimizer_prompt_failed"
	FailureKindOptimizerFailed              FailureKind = "optimizer_failed"
	FailureKindOptimizerToolFailed          FailureKind = "optimizer_tool_failed"
	FailureKindPolicyProposalFailed         FailureKind = "policy_proposal_failed"
	FailureKindPolicyStageWriteFailed       FailureKind = "policy_stage_write_failed"
	FailureKindPolicyPipelineFailed         FailureKind = "policy_pipeline_failed"
	FailureKindPolicyPipelineInfrastructure FailureKind = "policy_pipeline_infrastructure_failed"
	FailureKindAcceptProposalFailed         FailureKind = "accept_proposal_failed"
	FailureKindOptimizerRetriesExhausted    FailureKind = "optimizer_retries_exhausted"
	FailureKindContextCancelled             FailureKind = "context_cancelled"
)

// AttemptState is one optimizer-attempt outcome.
type AttemptState string

const (
	AttemptStatePending           AttemptState = "pending"
	AttemptStatePromptRendered    AttemptState = "prompt_rendered"
	AttemptStateProposalFinalized AttemptState = "proposal_finalized"
	AttemptStatePipelineFailed    AttemptState = "pipeline_failed"
	AttemptStateAccepted          AttemptState = "accepted"
	AttemptStateFailed            AttemptState = "failed"
)

// RetryPolicy controls bounded optimizer retries across attempts.
type RetryPolicy struct {
	MaxAttempts                  int
	RetryOnModelError            bool
	RetryOnToolFailure           bool
	RetryOnFinalizationFailure   bool
	RetryOnPolicyPipelineFailure bool
	RetryOnInfrastructureFailure bool
}

// DefaultRetryPolicy returns the first optimizer retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:                  3,
		RetryOnModelError:            true,
		RetryOnToolFailure:           true,
		RetryOnFinalizationFailure:   true,
		RetryOnPolicyPipelineFailure: true,
		RetryOnInfrastructureFailure: false,
	}
}

// ModelConfig records the model choice for one optimizer agent.
type ModelConfig struct {
	Provider        string `json:"provider"`
	Name            string `json:"name"`
	MaxOutputTokens int    `json:"max_output_tokens,omitempty"`
}

// Bounds records optimizer runtime bounds from the manifest.
type Bounds struct {
	MaxModelTurns  int `json:"max_model_turns,omitempty"`
	MaxToolCalls   int `json:"max_tool_calls,omitempty"`
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

// ToolPolicy records the optimizer agent allow/deny tool policy.
type ToolPolicy struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

// AgentConfig is the provider-free optimizer agent configuration.
type AgentConfig struct {
	Model        ModelConfig `json:"model"`
	Bounds       Bounds      `json:"bounds"`
	Tools        ToolPolicy  `json:"tools,omitempty"`
	SystemPrompt string      `json:"system_prompt,omitempty"`
}

// Target identifies the policy artifact being optimized.
type Target struct {
	InputArtifactID  domain.ArtifactID `json:"input_artifact_id"`
	OutputArtifactID domain.ArtifactID `json:"output_artifact_id"`
	OutputName       string            `json:"output_name"`
	InterfaceID      string            `json:"interface_id"`
}

// ParentRunRef identifies the completed evaluation bundle used as evidence.
type ParentRunRef struct {
	ArtifactID domain.ArtifactID `json:"artifact_id"`
	BundleID   string            `json:"bundle_id,omitempty"`
	BundlePath domain.HostPath   `json:"bundle_path,omitempty"`
}

// ReportSummary is the prompt-safe parent-run report projection.
type ReportSummary struct {
	ReportID          domain.ReportID     `json:"report_id"`
	Decision          string              `json:"decision,omitempty"`
	DecisionReason    string              `json:"decision_reason,omitempty"`
	CandidateSystemID domain.SystemID     `json:"candidate_system_id,omitempty"`
	CandidatePolicyID domain.PolicyID     `json:"candidate_policy_id,omitempty"`
	BaselineUsage     score.UsageEvidence `json:"baseline_usage,omitempty"`
	CandidateUsage    score.UsageEvidence `json:"candidate_usage,omitempty"`
	Comparisons       []MetricSummary     `json:"comparisons,omitempty"`
}

// MetricSummary is one report-safe metric delta summary.
type MetricSummary struct {
	Metric    string  `json:"metric"`
	Baseline  float64 `json:"baseline"`
	Candidate float64 `json:"candidate"`
	Delta     float64 `json:"delta"`
}

// PolicySource is one concrete policy artifact source.
type PolicySource struct {
	ArtifactID  domain.ArtifactID `json:"artifact_id"`
	Path        string            `json:"path,omitempty"`
	InterfaceID string            `json:"interface_id"`
	Source      string            `json:"source"`
}

// Evidence is the prompt-safe optimizer evidence payload.
type Evidence struct {
	ParentRun       ParentRunRef                 `json:"parent_run"`
	IncludedKinds   []string                     `json:"included_kinds,omitempty"`
	DeniedKinds     []string                     `json:"denied_kinds,omitempty"`
	ReportSummary   *ReportSummary               `json:"report_summary,omitempty"`
	ScoreEvidence   *score.ScoreEvidenceDocument `json:"score_evidence,omitempty"`
	ObjectiveResult *score.ObjectiveResult       `json:"objective_result,omitempty"`
	InputPolicy     PolicySource                 `json:"input_policy"`
}

// Spec is one resolved optimizer execution request.
type Spec struct {
	Target   Target      `json:"target"`
	Agent    AgentConfig `json:"agent"`
	Evidence Evidence    `json:"evidence"`
}

// Proposal is one finalized replacement policy artifact proposal.
type Proposal struct {
	ArtifactID   domain.ArtifactID `json:"artifact_id"`
	ArtifactName string            `json:"artifact_name"`
	InterfaceID  string            `json:"interface_id"`
	Code         string            `json:"code"`
	Summary      string            `json:"summary,omitempty"`
	RiskNotes    []string          `json:"risk_notes,omitempty"`
}

// Failure is the typed failed optimizer outcome.
type Failure struct {
	Phase            Phase       `json:"phase"`
	Kind             FailureKind `json:"kind"`
	Message          string      `json:"message"`
	Attempt          int         `json:"attempt,omitempty"`
	Retryable        bool        `json:"retryable,omitempty"`
	PipelineCategory string      `json:"pipeline_category,omitempty"`
	PipelineFeedback string      `json:"pipeline_feedback,omitempty"`
	Cause            error       `json:"-"`
}

func (f *Failure) Error() string {
	if f == nil {
		return "<nil>"
	}
	if f.Cause == nil {
		return fmt.Sprintf("%s/%s: %s", f.Phase, f.Kind, f.Message)
	}
	return fmt.Sprintf("%s/%s: %s: %v", f.Phase, f.Kind, f.Message, f.Cause)
}

// Unwrap exposes the underlying cause.
func (f *Failure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Cause
}

// Attempt records one optimizer attempt outcome.
type Attempt struct {
	Number                 int                      `json:"number"`
	State                  AttemptState             `json:"state"`
	RenderedPrompt         string                   `json:"rendered_prompt,omitempty"`
	RawOutput              string                   `json:"raw_output,omitempty"`
	Proposal               *Proposal                `json:"proposal,omitempty"`
	Failure                *Failure                 `json:"failure,omitempty"`
	PipelineResults        []pipeline.StepResult    `json:"-"`
	PipelineClassification *pipeline.Classification `json:"-"`
	RetryFeedback          string                   `json:"retry_feedback,omitempty"`
}

// Result is the typed outcome for one optimizer run.
type Result struct {
	Success        bool      `json:"success"`
	Proposal       *Proposal `json:"proposal,omitempty"`
	Failure        *Failure  `json:"failure,omitempty"`
	Attempts       []Attempt `json:"attempts,omitempty"`
	Phases         []Phase   `json:"phases,omitempty"`
	RenderedPrompt string    `json:"rendered_prompt,omitempty"`
	RawOutput      string    `json:"raw_output,omitempty"`
}
