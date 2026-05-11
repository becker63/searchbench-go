// Package fake is a deterministic Scorer used by local-only fake rounds.
//
// It satisfies the structural compare.Scorer interface without binding to that
// internal package, so rounds compose it as the default scorer when no real
// scoring backend is configured.
package fake

import (
	"context"

	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Scorer emits a deterministic ScoreSet that prefers the iterative-context
// challenger backend over the conservative incumbent. It is the canonical
// stand-in scorer for local-fake rounds.
type Scorer struct{}

// New constructs a Scorer.
func New() Scorer { return Scorer{} }

// Score satisfies the structural compare.Scorer interface.
func (Scorer) Score(_ context.Context, input score.Input) (score.ScoreSet, error) {
	if input.Run.Spec().System.Backend == domain.BackendIterativeContext {
		return score.NewScoreSet(
			score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 2},
			score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 3},
			score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.85},
			score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.15},
			score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.90},
		)
	}
	return score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 6},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 7},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.35},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.60},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.30},
	)
}
