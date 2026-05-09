package compare

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

type Plan struct {
	Systems domain.Pair[domain.SystemSpec]
	Matches domain.NonEmpty[domain.MatchSpec]
}

// NewPlan constructs an executable baseline/candidate comparison request.
func NewPlan(
	systems domain.Pair[domain.SystemSpec],
	tasks domain.NonEmpty[domain.MatchSpec],
) Plan {
	return Plan{
		Systems: systems,
		Matches: tasks,
	}
}

// Validate checks the executable comparison inputs before orchestration starts.
//
// It requires a non-empty task slice plus individually valid baseline and
// candidate systems, wrapping system errors with their comparison role.
func (p Plan) Validate() error {
	if err := p.Matches.Validate(); err != nil {
		return err
	}
	if err := p.Systems.Incumbent.Validate(); err != nil {
		return fmt.Errorf("baseline system: %w", err)
	}
	if err := p.Systems.Challenger.Validate(); err != nil {
		return fmt.Errorf("candidate system: %w", err)
	}
	return nil
}

// ReportSpec converts the executable plan into its report-safe identity.
//
// The resulting ComparisonSpec contains SystemRef values rather than full
// executable SystemSpec values so reports do not carry policy source by
// default.
func (p Plan) ReportSpec() report.ComparisonSpec {
	return report.NewComparisonSpec(p.Systems, p.Matches)
}
