package round

import (
	"context"
	"errors"

	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

const graphHopExplanation = "call-graph hops via tree-sitter index (MinCallHopsAcrossFiles)"

// LocalizationGraphScorer recomputes hop-distance metrics from the indexed graph
// when predictions and oracle gold files participate in a non-empty call graph.
// Other metrics are preserved from Fallback (typically [evaluatorfake.Scorer]).
type LocalizationGraphScorer struct {
	Fallback compare.Scorer
}

// NewLocalizationGraphScorer wraps the fallback scorer.
func NewLocalizationGraphScorer(fallback compare.Scorer) LocalizationGraphScorer {
	return LocalizationGraphScorer{Fallback: fallback}
}

// Score satisfies [compare.Scorer].
func (s LocalizationGraphScorer) Score(ctx context.Context, input score.Input) (score.ScoreSet, error) {
	if s.Fallback == nil {
		return score.ScoreSet{}, errors.New("round: LocalizationGraphScorer requires non-nil Fallback")
	}
	base, err := s.Fallback.Score(ctx, input)
	if err != nil {
		return score.ScoreSet{}, err
	}
	pred := input.Run.Prediction.Files
	gold := input.Run.Spec().Match.Oracle.GoldFiles
	if len(pred) == 0 || len(gold) == 0 || input.Graph == nil {
		return base, nil
	}
	g := input.Graph
	gh, okG := codegraph.MinCallHopsAcrossFiles(g, pred, gold)
	ih, okI := codegraph.MinCallHopsAcrossFiles(g, gold, pred)
	if !okG && !okI {
		return base, nil
	}
	var goldHop, issueHop score.HopDistance
	switch {
	case okG && okI:
		goldHop = score.HopDistance(gh)
		issueHop = score.HopDistance(ih)
	case okG:
		v := score.HopDistance(gh)
		goldHop, issueHop = v, v
	default:
		v := score.HopDistance(ih)
		goldHop, issueHop = v, v
	}
	return score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: goldHop, Explanation: graphHopExplanation},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: issueHop, Explanation: graphHopExplanation},
		base.TokenEfficiency,
		base.Cost,
		base.Composite,
	)
}
