package bundlefs

import (
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

const (
	schemaVersion         = "searchbench.bundle.v1"
	defaultRenderedReport = "round-report.txt"
	completeMarkerName    = "COMPLETE"
)

// RenderedReport is the optional human-readable bundle artifact.
type RenderedReport struct {
	FileName  string
	MediaType string
	Content   string
}

// RoundBundleInput is the caller-supplied round bundle write contract.
type RoundBundleInput struct {
	RootPath        domain.HostPath
	BundleID        string
	ResolvedInput   any
	RoundReport     report.RoundReport
	RoundEvidence   score.RoundEvidenceDocument
	ObjectiveResult *score.ObjectiveResult
	RenderedReport  *RenderedReport
	CreatedAt       time.Time
}

// RoundBundleRef identifies one completed immutable round bundle on local disk.
type RoundBundleRef struct {
	BundleID  string
	Path      domain.HostPath
	Files     []BundleFile
	CreatedAt time.Time
}

// BundleRequest is a transitional alias for RoundBundleInput.
//
// TODO(issue-32): remove after callers use RoundBundleInput directly.
type BundleRequest = RoundBundleInput

// BundleRef is a transitional alias for RoundBundleRef.
//
// TODO(issue-32): remove after callers use RoundBundleRef directly.
type BundleRef = RoundBundleRef

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
