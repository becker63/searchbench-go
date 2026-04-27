package artifact

import (
	"time"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/score"
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

// ScoreEvidence is the report-derived score projection written into score.json.
type ScoreEvidence struct {
	PromotionDecision report.PromotionDecision `json:"promotion_decision"`
	RunCounts         RoleCounts               `json:"run_counts"`
	FailureCounts     RoleCounts               `json:"failure_counts"`
	Metrics           []MetricEvidence         `json:"metrics"`
	Regressions       []report.Regression      `json:"regressions,omitempty"`
}

// RoleCounts records baseline/candidate counts in stable role order.
type RoleCounts struct {
	Baseline  int `json:"baseline"`
	Candidate int `json:"candidate"`
}

// MetricEvidence preserves the metric-level comparison shape already present in
// the report, plus canonical direction/improvement flags from the score
// package.
type MetricEvidence struct {
	Metric    score.MetricName `json:"metric"`
	Direction score.Direction  `json:"direction,omitempty"`
	Baseline  float64          `json:"baseline"`
	Candidate float64          `json:"candidate"`
	Delta     float64          `json:"delta"`
	Improved  bool             `json:"improved"`
	Regressed bool             `json:"regressed"`
}
