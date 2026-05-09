package evaluation

import (
	"time"

	"github.com/becker63/searchbench-go/internal/app/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// ResolveRequest configures one manifest-driven evaluation plan resolution.
type ResolveRequest struct {
	ManifestPath       string
	BundleRootOverride string
	BundleID           string
	ReportID           domain.ReportID
	ParentRef          *score.ObjectiveEvidenceRef
	ParentScorePath    string
	Now                func() time.Time
}

// Plan is the canonical app-layer evaluation projection used by execution.
type Plan struct {
	ManifestPath   string                           `json:"manifest_path,omitempty"`
	ExperimentName string                           `json:"experiment_name,omitempty"`
	Mode           string                           `json:"mode,omitempty"`
	Dataset        DatasetConfig                    `json:"dataset,omitempty"`
	Systems        domain.Pair[domain.SystemSpec]   `json:"-"`
	Matches        domain.NonEmpty[domain.MatchSpec] `json:"matches"`
	Parallelism    compare.Parallelism              `json:"-"`
	Evaluator      EvaluatorConfig                  `json:"evaluator,omitempty"`
	Scoring        ScoringConfig                    `json:"scoring,omitempty"`
	Output         OutputConfig                     `json:"output,omitempty"`
	Report         ReportConfig                     `json:"report_options,omitempty"`
	Bundle         BundleConfig                     `json:"bundle,omitempty"`
	ReportID       domain.ReportID                  `json:"report_id,omitempty"`
	CreatedAt      time.Time                        `json:"created_at"`
}

// ComparePlan converts the canonical evaluation model into the executable
// comparison plan used by current runners.
func (p Plan) ComparePlan() compare.Plan {
	return compare.NewPlan(p.Systems, p.Matches)
}

// BundleSystems returns the report-safe system identities used for serialization.
func (p Plan) BundleSystems() domain.Pair[domain.SystemRef] {
	return p.ComparePlan().ReportSpec().Systems
}

// DatasetConfig records the resolved dataset selection from the manifest.
type DatasetConfig struct {
	Kind     string `json:"kind,omitempty"`
	Name     string `json:"name,omitempty"`
	Config   string `json:"config,omitempty"`
	Split    string `json:"split,omitempty"`
	MaxItems *int   `json:"max_items,omitempty"`
}

// EvaluatorConfig records the resolved evaluator model, bounds, and retry
// policy projected from the manifest.
type EvaluatorConfig struct {
	Model  EvaluatorModelConfig  `json:"model,omitempty"`
	Bounds EvaluatorBoundsConfig `json:"bounds,omitempty"`
	Retry  RetryPolicyConfig     `json:"retry,omitempty"`
}

// EvaluatorModelConfig records the resolved evaluator model choice.
type EvaluatorModelConfig struct {
	Provider        string `json:"provider,omitempty"`
	Name            string `json:"name,omitempty"`
	MaxOutputTokens int    `json:"max_output_tokens,omitempty"`
}

// EvaluatorBoundsConfig records the resolved evaluator bounds.
type EvaluatorBoundsConfig struct {
	MaxModelTurns  int `json:"max_model_turns,omitempty"`
	MaxToolCalls   int `json:"max_tool_calls,omitempty"`
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`
}

// RetryPolicyConfig records the resolved evaluator retry policy.
type RetryPolicyConfig struct {
	MaxAttempts                int  `json:"max_attempts,omitempty"`
	RetryOnModelError          bool `json:"retry_on_model_error,omitempty"`
	RetryOnToolFailure         bool `json:"retry_on_tool_failure,omitempty"`
	RetryOnFinalizationFailure bool `json:"retry_on_finalization_failure,omitempty"`
	RetryOnInvalidPrediction   bool `json:"retry_on_invalid_prediction,omitempty"`
}

// ScoringConfig records the resolved scoring objective and evidence refs.
type ScoringConfig struct {
	ObjectivePath   string                      `json:"objective_path,omitempty"`
	CurrentEvidence score.ObjectiveEvidenceRef  `json:"current_evidence"`
	ParentEvidence  *score.ObjectiveEvidenceRef `json:"parent_evidence,omitempty"`
	ParentScorePath string                      `json:"-"`
}

// OutputConfig records resolved artifact output and policy path preferences.
type OutputConfig struct {
	BundleCollectionPath domain.HostPath     `json:"bundle_root,omitempty"`
	BundleWriterRoot     domain.HostPath     `json:"bundle_writer_root,omitempty"`
	ExpectedBundlePath   domain.HostPath     `json:"-"`
	ReportFormats        []string            `json:"report_formats,omitempty"`
	RenderHumanReport    bool                `json:"render_human_report,omitempty"`
	ResolvedPolicyPaths  ResolvedPolicyPaths `json:"resolved_policy_path,omitempty"`
}

// ResolvedPolicyPaths records resolved manifest-relative policy paths.
type ResolvedPolicyPaths struct {
	Incumbent  string `json:"incumbent,omitempty"`
	Challenger string `json:"challenger,omitempty"`
}

// ReportConfig records durable report rendering preferences.
type ReportConfig struct {
	Formats []string `json:"formats,omitempty"`
}

// BundleConfig records the resolved output bundle target.
type BundleConfig struct {
	ID string `json:"id,omitempty"`
}
