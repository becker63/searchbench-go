package report

import (
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// RoundReport is the central SearchBench round artifact.
//
// It compares two policies over the same match slice and records whether the
// challenger should advance, be reviewed, or be rejected.
type RoundReport struct {
	ID        domain.ReportID `json:"id"`
	CreatedAt time.Time       `json:"created_at"`

	Spec ComparisonSpec `json:"spec"`

	Runs     domain.Pair[[]score.ScoredRun] `json:"runs"`
	Failures domain.Pair[[]run.RunFailure]  `json:"failures,omitempty"`

	Comparisons []ScoreComparison `json:"comparisons"`
	Regressions []Regression      `json:"regressions"`
	Decision    Decision          `json:"decision"`

	Artifacts []domain.ReportArtifactRef `json:"artifacts,omitempty"`
}

// NewRoundReport constructs the central SearchBench round artifact from
// already-compared executions and failures.
func NewRoundReport(
	id domain.ReportID,
	spec ComparisonSpec,
	runs domain.Pair[[]score.ScoredRun],
	failures domain.Pair[[]run.RunFailure],
	comparisons []ScoreComparison,
	regressions []Regression,
	decision Decision,
) RoundReport {
	return RoundReport{
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
