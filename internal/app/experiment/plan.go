package experiment

import (
	"time"

	"github.com/becker63/searchbench-go/internal/app/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Request configures one manifest-driven resolved experiment.
type Request struct {
	ManifestPath        string
	BundleRootOverride  string
	BundleID            string
	ReportID            domain.ReportID
	ParentRef           *score.ObjectiveEvidenceRef
	ParentScorePath     string
	Now                 func() time.Time
	PklCommand          []string
	DisableRenderReport bool
}

// ResolvedExperiment is the canonical app-layer manifest projection used by
// execution strategies such as localrun.
type ResolvedExperiment struct {
	ManifestPath   string
	ExperimentName string
	Mode           string
	Dataset        DatasetConfig
	Systems        domain.Pair[domain.SystemSpec]
	Tasks          domain.NonEmpty[domain.TaskSpec]
	Parallelism    compare.Parallelism
	Evaluator      EvaluatorConfig
	Scoring        ScoringConfig
	Output         OutputConfig
	Report         ReportConfig
	Bundle         BundleConfig
	ReportID       domain.ReportID
	CreatedAt      time.Time
}

// ComparePlan converts the canonical experiment model into the executable
// comparison plan used by current runners.
func (r ResolvedExperiment) ComparePlan() compare.Plan {
	return compare.NewPlan(r.Systems, r.Tasks)
}

// DatasetConfig records the resolved dataset selection from the manifest.
type DatasetConfig struct {
	Kind     string
	Name     string
	Config   string
	Split    string
	MaxItems *int
}

// EvaluatorConfig records the resolved evaluator model, bounds, and retry
// policy projected from the manifest.
type EvaluatorConfig struct {
	Model  EvaluatorModelConfig
	Bounds EvaluatorBoundsConfig
	Retry  RetryPolicyConfig
}

// EvaluatorModelConfig records the resolved evaluator model choice.
type EvaluatorModelConfig struct {
	Provider        string
	Name            string
	MaxOutputTokens int
}

// EvaluatorBoundsConfig records the resolved evaluator bounds.
type EvaluatorBoundsConfig struct {
	MaxModelTurns  int
	MaxToolCalls   int
	TimeoutSeconds int
}

// RetryPolicyConfig records the resolved evaluator retry policy.
type RetryPolicyConfig struct {
	MaxAttempts                int
	RetryOnModelError          bool
	RetryOnToolFailure         bool
	RetryOnFinalizationFailure bool
	RetryOnInvalidPrediction   bool
}

// ScoringConfig records the resolved scoring objective and evidence refs.
type ScoringConfig struct {
	ObjectivePath   string
	CurrentEvidence score.ObjectiveEvidenceRef
	ParentEvidence  *score.ObjectiveEvidenceRef
	ParentScorePath string
}

// OutputConfig records resolved artifact output and policy path preferences.
type OutputConfig struct {
	BundleCollectionPath domain.HostPath
	BundleWriterRoot     domain.HostPath
	ExpectedBundlePath   domain.HostPath
	ReportFormat         string
	RenderHumanReport    bool
	ResolvedPolicyPaths  ResolvedPolicyPaths
}

// ResolvedPolicyPaths records resolved manifest-relative policy paths.
type ResolvedPolicyPaths struct {
	Baseline  string
	Candidate string
}

// ReportConfig records durable report rendering preferences.
type ReportConfig struct {
	Format string
}

// BundleConfig records the resolved output bundle target.
type BundleConfig struct {
	ID string
}
