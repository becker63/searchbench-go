package localrun

import (
	artifact "github.com/becker63/searchbench-go/internal/adapters/artifact/fsbundle"
	appExperiment "github.com/becker63/searchbench-go/internal/app/experiment"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Request configures one local fake manifest-driven run.
type Request = appExperiment.Request

// Result is the completed local fake manifest-driven run.
type Result struct {
	ManifestPath    string
	Bundle          artifact.BundleRef
	ReportID        domain.ReportID
	CandidateReport report.CandidateReport
	ScoreEvidence   score.ScoreEvidenceDocument
	ObjectiveResult *score.ObjectiveResult
}
