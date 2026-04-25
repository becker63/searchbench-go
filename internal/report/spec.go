package report

import (
	"errors"
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
)

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

func (s ComparisonSpec) Validate() error {
	if err := s.Tasks.Validate(); err != nil {
		return err
	}
	if s.Systems.Baseline.ID.Empty() {
		return errors.New("baseline system id is required")
	}
	if s.Systems.Candidate.ID.Empty() {
		return errors.New("candidate system id is required")
	}
	if s.Systems.Baseline.Fingerprint == "" {
		return fmt.Errorf("baseline system fingerprint is required")
	}
	if s.Systems.Candidate.Fingerprint == "" {
		return fmt.Errorf("candidate system fingerprint is required")
	}
	return nil
}
