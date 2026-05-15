package report

import (
	"errors"
	"fmt"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// ComparisonSpec declares the policies and matches being compared.
//
// This is the planned report boundary:
//
//	incumbent policy + challenger policy + fixed task slice
type ComparisonSpec struct {
	Policies domain.Pair[domain.SystemRef]     `json:"policies"`
	Matches  domain.NonEmpty[domain.MatchSpec] `json:"matches"`
}

// NewComparisonSpec constructs a report-safe comparison boundary from full
// executable policies.
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
		Policies: systems,
		Matches:  tasks,
	}
}

// Validate checks that the report boundary is structurally meaningful.
func (s ComparisonSpec) Validate() error {
	if err := s.Matches.Validate(); err != nil {
		return err
	}
	if s.Policies.Incumbent.ID.Empty() {
		return errors.New("incumbent policy id is required")
	}
	if s.Policies.Challenger.ID.Empty() {
		return errors.New("challenger policy id is required")
	}
	if s.Policies.Incumbent.Fingerprint == "" {
		return fmt.Errorf("incumbent policy fingerprint is required")
	}
	if s.Policies.Challenger.Fingerprint == "" {
		return fmt.Errorf("challenger policy fingerprint is required")
	}
	return nil
}
