package report

import (
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// RoleExecutionOutcome captures one policy role outcome for a single match.
type RoleExecutionOutcome struct {
	Scored *score.ScoredRun
	Failed *run.RunFailure
}

// NewRoleExecutionOutcome constructs an outcome from mutually exclusive pointers.
func NewRoleExecutionOutcome(scored *score.ScoredRun, failed *run.RunFailure) RoleExecutionOutcome {
	return RoleExecutionOutcome{Scored: scored, Failed: failed}
}

// MatchExecutionRecord aligns one planned match with incumbent/challenger outcomes.
//
// It preserves task index order via construction from ordered compare.TaskWorkResult
// slices and is the durable seam for downstream localization evidence builders.
type MatchExecutionRecord struct {
	MatchID     domain.MatchID       `json:"match_id"`
	Task        domain.MatchSpec     `json:"task"`
	Incumbent   RoleExecutionOutcome `json:"incumbent"`
	Challenger  RoleExecutionOutcome `json:"challenger"`
	Regressions []Regression         `json:"regressions,omitempty"`
}
