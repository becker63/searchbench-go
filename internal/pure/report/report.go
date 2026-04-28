package report

import (
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/run"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// CandidateReport is the central Searchbench product object.
//
// It compares two systems over the same task slice and records whether the
// candidate should promote, be reviewed, or be rejected.
type CandidateReport struct {
	ID        domain.ReportID `json:"id"`
	CreatedAt time.Time       `json:"created_at"`

	Spec ComparisonSpec `json:"spec"`

	Runs     domain.Pair[[]score.ScoredRun] `json:"runs"`
	Failures domain.Pair[[]run.RunFailure]  `json:"failures,omitempty"`

	Comparisons []ScoreComparison `json:"comparisons"`
	Regressions []Regression      `json:"regressions"`
	Decision    PromotionDecision `json:"decision"`

	Artifacts []domain.ReportArtifactRef `json:"artifacts,omitempty"`
}

// NewCandidateReport constructs the central Searchbench release artifact from
// already-compared runs and failures.
func NewCandidateReport(
	id domain.ReportID,
	spec ComparisonSpec,
	runs domain.Pair[[]score.ScoredRun],
	failures domain.Pair[[]run.RunFailure],
	comparisons []ScoreComparison,
	regressions []Regression,
	decision PromotionDecision,
) CandidateReport {
	return CandidateReport{
		ID:          id,
		CreatedAt:   time.Now().UTC(),
		Spec:        spec,
		Runs:        runs,
		Failures:    failures,
		Comparisons: comparisons,
		Regressions: regressions,
		Decision:    decision,
	}
}
