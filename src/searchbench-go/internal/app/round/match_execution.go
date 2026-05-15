package round

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

func matchExecutionRecordsFromTaskWork(matches domain.NonEmpty[domain.MatchSpec], work []compare.TaskWorkResult) ([]report.MatchExecutionRecord, error) {
	if len(work) != matches.Len() {
		return nil, fmt.Errorf("match execution records: task work len %d != matches len %d", len(work), matches.Len())
	}
	sorted := slices.Clone(work)
	slices.SortFunc(sorted, func(a, b compare.TaskWorkResult) int {
		return cmp.Compare(a.Index, b.Index)
	})
	all := matches.All()
	out := make([]report.MatchExecutionRecord, 0, len(sorted))
	for i, tw := range sorted {
		if tw.Index != i {
			return nil, fmt.Errorf("match execution records: unexpected task index %d at position %d", tw.Index, i)
		}
		if tw.MatchID != all[i].ID {
			return nil, fmt.Errorf("match execution records: match id %s != plan match id %s", tw.MatchID, all[i].ID)
		}
		tr := tw.Result
		out = append(out, report.MatchExecutionRecord{
			MatchID: tw.MatchID,
			Task:    all[i],
			Incumbent: report.NewRoleExecutionOutcome(
				tr.Runs.Incumbent,
				tr.Failures.Incumbent,
			),
			Challenger: report.NewRoleExecutionOutcome(
				tr.Runs.Challenger,
				tr.Failures.Challenger,
			),
			Regressions: append([]report.Regression(nil), tr.Regressions...),
		})
	}
	return out, nil
}
