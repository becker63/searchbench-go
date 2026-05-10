package compare

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

type Plan struct {
	Policies domain.Pair[domain.SystemSpec]
	Matches  domain.NonEmpty[domain.MatchSpec]
}

// NewPlan constructs an executable incumbent/challenger comparison request.
func NewPlan(
	systems domain.Pair[domain.SystemSpec],
	tasks domain.NonEmpty[domain.MatchSpec],
) Plan {
	return Plan{
		Policies: systems,
		Matches:  tasks,
	}
}

// Validate checks the executable comparison inputs before orchestration starts.
//
// It requires a non-empty task slice plus individually valid incumbent and
// challenger policies, wrapping validation errors with their comparison role.
func (p Plan) Validate() error {
	if err := p.Matches.Validate(); err != nil {
		return err
	}
	if err := p.Policies.Incumbent.Validate(); err != nil {
		return fmt.Errorf("incumbent policy: %w", err)
	}
	if err := p.Policies.Challenger.Validate(); err != nil {
		return fmt.Errorf("challenger policy: %w", err)
	}
	return nil
}

// ReportSpec converts the executable plan into its report-safe identity.
//
// The resulting ComparisonSpec contains SystemRef values rather than full
// executable SystemSpec values so reports do not carry policy source by
// default.
func (p Plan) ReportSpec() report.ComparisonSpec {
	return report.NewComparisonSpec(p.Policies, p.Matches)
}
