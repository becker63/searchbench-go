package report

import (
	"errors"
	"fmt"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// ComparisonSpec declares the systems and matches being compared.
//
// This is the planned report boundary:
//
//	baseline system + candidate system + fixed task slice
type ComparisonSpec struct {
	Systems domain.Pair[domain.SystemRef]    `json:"systems"`
	Matches domain.NonEmpty[domain.MatchSpec] `json:"matches"`
}

// NewComparisonSpec constructs a report-safe comparison boundary from full
// executable systems.
//
// The returned spec stores SystemRef values rather than SystemSpec values so
// reports do not carry policy source.
func NewComparisonSpec(
	systems domain.Pair[domain.SystemSpec],
	tasks domain.NonEmpty[domain.MatchSpec],
) ComparisonSpec {
	return NewComparisonSpecFromRefs(domain.NewPair(systems.Incumbent.Ref(), systems.Challenger.Ref()), tasks)
}

// NewComparisonSpecFromRefs constructs a report-safe comparison boundary from
// report-safe system identities.
func NewComparisonSpecFromRefs(
	systems domain.Pair[domain.SystemRef],
	tasks domain.NonEmpty[domain.MatchSpec],
) ComparisonSpec {
	return ComparisonSpec{
		Systems: systems,
		Matches: tasks,
	}
}

// Validate checks that the report boundary is structurally meaningful.
func (s ComparisonSpec) Validate() error {
	if err := s.Matches.Validate(); err != nil {
		return err
	}
	if s.Systems.Incumbent.ID.Empty() {
		return errors.New("incumbent system id is required")
	}
	if s.Systems.Challenger.ID.Empty() {
		return errors.New("challenger system id is required")
	}
	if s.Systems.Incumbent.Fingerprint == "" {
		return fmt.Errorf("incumbent system fingerprint is required")
	}
	if s.Systems.Challenger.Fingerprint == "" {
		return fmt.Errorf("challenger system fingerprint is required")
	}
	return nil
}
