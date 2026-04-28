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
	Systems        domain.Pair[domain.SystemRef]    `json:"systems"`
	Tasks          domain.NonEmpty[domain.TaskSpec] `json:"tasks"`
	Parallelism    ParallelismConfig                `json:"parallelism,omitempty"`
	ScoringProfile string                           `json:"scoring_profile,omitempty"`
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
