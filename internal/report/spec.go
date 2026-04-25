package report

import "github.com/becker63/searchbench-go/internal/domain"

// ComparisonSpec declares the systems and tasks being compared.
//
// This is the planned report boundary:
//
//	baseline system + candidate system + fixed task slice
type ComparisonSpec struct {
	Systems domain.Pair[domain.SystemRef]    `json:"systems"`
	Tasks   domain.NonEmpty[domain.TaskSpec] `json:"tasks"`
}

func NewComparisonSpec(
	systems domain.Pair[domain.SystemSpec],
	tasks domain.NonEmpty[domain.TaskSpec],
) ComparisonSpec {
	return NewComparisonSpecFromRefs(domain.NewPair(systems.Baseline.Ref(), systems.Candidate.Ref()), tasks)
}

func NewComparisonSpecFromRefs(
	systems domain.Pair[domain.SystemRef],
	tasks domain.NonEmpty[domain.TaskSpec],
) ComparisonSpec {
	return ComparisonSpec{
		Systems: systems,
		Tasks:   tasks,
	}
}
