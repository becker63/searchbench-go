package bundlefs

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
	ResolvedInput   any
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
