package compare

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
)

type Plan struct {
	Systems domain.Pair[domain.SystemSpec]
	Tasks   domain.NonEmpty[domain.TaskSpec]
}

func NewPlan(
	systems domain.Pair[domain.SystemSpec],
	tasks domain.NonEmpty[domain.TaskSpec],
) Plan {
	return Plan{
		Systems: systems,
		Tasks:   tasks,
	}
}

func (p Plan) Validate() error {
	if err := p.Tasks.Validate(); err != nil {
		return err
	}
	if err := p.Systems.Baseline.Validate(); err != nil {
		return fmt.Errorf("baseline system: %w", err)
	}
	if err := p.Systems.Candidate.Validate(); err != nil {
		return fmt.Errorf("candidate system: %w", err)
	}
	return nil
}

func (p Plan) ReportSpec() report.ComparisonSpec {
	return report.NewComparisonSpec(p.Systems, p.Tasks)
}
