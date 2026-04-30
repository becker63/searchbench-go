package artifact

import (
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

const (
	schemaVersion         = "searchbench.bundle.v1"
	defaultRenderedReport = "report.txt"
	completeMarkerName    = "COMPLETE"
)

// ResolvedComparisonInput is the first canonical bundle input document.
//
// It is intentionally bundle-local for now: enough structure to explain what
// was compared without moving orchestration models into the artifact package.
type ResolvedComparisonInput struct {
	ManifestPath   string                           `json:"manifest_path,omitempty"`
	ExperimentName string                           `json:"experiment_name,omitempty"`
	Mode           string                           `json:"mode,omitempty"`
	Dataset        DatasetConfig                    `json:"dataset,omitempty"`
	Systems        domain.Pair[domain.SystemRef]    `json:"systems"`
	Tasks          domain.NonEmpty[domain.TaskSpec] `json:"tasks"`
	Parallelism    ParallelismConfig                `json:"parallelism,omitempty"`
	Evaluator      EvaluatorConfig                  `json:"evaluator,omitempty"`
	Scoring        ScoringConfig                    `json:"scoring,omitempty"`
	Output         OutputConfig                     `json:"output,omitempty"`
	ReportOptions  ReportOptions                    `json:"report_options,omitempty"`
}

// ParallelismConfig records the resolved task-level comparison policy.
type ParallelismConfig struct {
	Mode       string `json:"mode,omitempty"`
	MaxWorkers int    `json:"max_workers,omitempty"`
	FailFast   bool   `json:"fail_fast,omitempty"`
}

// ReportOptions records durable report-output preferences without terminal
// styling details.
type ReportOptions struct {
	Format string `json:"format,omitempty"`
}

// DatasetConfig records the resolved dataset selection from the experiment
// manifest.
type DatasetConfig struct {
	Kind     string `json:"kind,omitempty"`
	Name     string `json:"name,omitempty"`
	Config   string `json:"config,omitempty"`
	Split    string `json:"split,omitempty"`
	MaxItems *int   `json:"max_items,omitempty"`
}

// EvaluatorConfig records the resolved evaluator model and bounds used to
// project executable runtime settings.
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

// RetryPolicyConfig records the resolved retry policy from the manifest.
type RetryPolicyConfig struct {
	MaxAttempts                int  `json:"max_attempts,omitempty"`
	RetryOnModelError          bool `json:"retry_on_model_error,omitempty"`
	RetryOnToolFailure         bool `json:"retry_on_tool_failure,omitempty"`
	RetryOnFinalizationFailure bool `json:"retry_on_finalization_failure,omitempty"`
	RetryOnInvalidPrediction   bool `json:"retry_on_invalid_prediction,omitempty"`
}

// ScoringConfig records the resolved scoring objective inputs.
type ScoringConfig struct {
	ObjectivePath string         `json:"objective_path,omitempty"`
	Evidence      EvidenceConfig `json:"evidence,omitempty"`
}

// EvidenceConfig records the durable evidence refs used for scoring.
type EvidenceConfig struct {
	Current score.ObjectiveEvidenceRef  `json:"current"`
	Parent  *score.ObjectiveEvidenceRef `json:"parent,omitempty"`
}

// OutputConfig records the resolved output/bundle preferences.
type OutputConfig struct {
	BundleRoot         string             `json:"bundle_root,omitempty"`
	BundleWriterRoot   string             `json:"bundle_writer_root,omitempty"`
	ReportFormat       string             `json:"report_format,omitempty"`
	RenderHumanReport  bool               `json:"render_human_report,omitempty"`
	ResolvedPolicyPath ResolvedPolicyPath `json:"resolved_policy_path,omitempty"`
}

// ResolvedPolicyPath records resolved manifest-relative policy artifact paths.
type ResolvedPolicyPath struct {
	Baseline  string `json:"baseline,omitempty"`
	Candidate string `json:"candidate,omitempty"`
}

// RenderedReport is the optional human-readable bundle artifact.
type RenderedReport struct {
	FileName  string
	MediaType string
	Content   string
}

// BundleRequest is the caller-supplied bundle write contract.
type BundleRequest struct {
	RootPath        domain.HostPath
	BundleID        string
	ResolvedInput   ResolvedComparisonInput
	CandidateReport report.CandidateReport
	ScoreEvidence   score.ScoreEvidenceDocument
	ObjectiveResult *score.ObjectiveResult
	RenderedReport  *RenderedReport
	CreatedAt       time.Time
}

// BundleRef identifies one completed immutable bundle on local disk.
type BundleRef struct {
	BundleID  string
	Path      domain.HostPath
	Files     []BundleFile
	CreatedAt time.Time
}

// BundleFile describes one serialized bundle artifact.
type BundleFile struct {
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	MediaType string `json:"media_type"`
	SHA256    string `json:"sha256,omitempty"`
}

// BundleMetadata is the deterministic inventory written into metadata.json.
type BundleMetadata struct {
	SchemaVersion string       `json:"schema_version"`
	BundleID      string       `json:"bundle_id"`
	CreatedAt     time.Time    `json:"created_at"`
	Files         []BundleFile `json:"files"`
}
