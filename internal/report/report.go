package report

import (
	"time"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/score"
)

// ComparisonSpec declares the systems and tasks being compared.
//
// This is the planned report boundary:
//
//	baseline system + candidate system + fixed task slice
type ComparisonSpec struct {
	Systems domain.Pair[domain.SystemSpec] `json:"systems"`
	Tasks   []domain.TaskSpec              `json:"tasks"`
}

// CandidateReport is the central Searchbench product object.
//
// It compares two systems over the same task slice and records whether the
// candidate should promote, be reviewed, or be rejected.
type CandidateReport struct {
	ID        domain.ReportID `json:"id"`
	CreatedAt time.Time       `json:"created_at"`

	Spec ComparisonSpec `json:"spec"`

	Runs domain.Pair[[]score.ScoredRun] `json:"runs"`

	Comparisons []ScoreComparison `json:"comparisons"`
	Regressions []Regression      `json:"regressions"`
	Decision    PromotionDecision `json:"decision"`

	Artifacts []domain.ReportArtifactRef `json:"artifacts,omitempty"`
}

// NewCandidateReport constructs a report from already-scored run sets.
func NewCandidateReport(
	id domain.ReportID,
	spec ComparisonSpec,
	runs domain.Pair[[]score.ScoredRun],
	comparisons []ScoreComparison,
	regressions []Regression,
	decision PromotionDecision,
) CandidateReport {
	return CandidateReport{
		ID:          id,
		CreatedAt:   time.Now().UTC(),
		Spec:        spec,
		Runs:        runs,
		Comparisons: comparisons,
		Regressions: regressions,
		Decision:    decision,
	}
}
