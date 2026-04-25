package score

type Direction string

const (
	// LowerIsBetter means smaller metric values are improvements.
	LowerIsBetter Direction = "lower_is_better"
	// HigherIsBetter means larger metric values are improvements.
	HigherIsBetter Direction = "higher_is_better"
)

// MetricDefinition binds one metric name to its improvement direction.
type MetricDefinition[T any] struct {
	Name      MetricName
	Direction Direction
}

// GoldHopDefinition is the canonical definition for gold_hop.
var GoldHopDefinition = MetricDefinition[HopDistance]{
	Name:      MetricGoldHop,
	Direction: LowerIsBetter,
}

// IssueHopDefinition is the canonical definition for issue_hop.
var IssueHopDefinition = MetricDefinition[HopDistance]{
	Name:      MetricIssueHop,
	Direction: LowerIsBetter,
}

// TokenEfficiencyDefinition is the canonical definition for token_efficiency.
var TokenEfficiencyDefinition = MetricDefinition[EfficiencyScore]{
	Name:      MetricTokenEfficiency,
	Direction: HigherIsBetter,
}

// CostDefinition is the canonical definition for cost.
var CostDefinition = MetricDefinition[CostScore]{
	Name:      MetricCost,
	Direction: LowerIsBetter,
}

// CompositeDefinition is the canonical definition for composite.
var CompositeDefinition = MetricDefinition[CompositeScore]{
	Name:      MetricComposite,
	Direction: HigherIsBetter,
}

// Improved reports whether the candidate value is better than the baseline
// under the given metric direction.
//
// Positive delta is not universally good; direction determines the meaning.
func Improved(def Direction, baseline, candidate float64) bool {
	switch def {
	case HigherIsBetter:
		return candidate > baseline
	case LowerIsBetter:
		return candidate < baseline
	default:
		return false
	}
}

// DirectionForMetric looks up the canonical improvement direction for a metric.
func DirectionForMetric(name MetricName) (Direction, bool) {
	switch name {
	case MetricGoldHop:
		return GoldHopDefinition.Direction, true
	case MetricIssueHop:
		return IssueHopDefinition.Direction, true
	case MetricTokenEfficiency:
		return TokenEfficiencyDefinition.Direction, true
	case MetricCost:
		return CostDefinition.Direction, true
	case MetricComposite:
		return CompositeDefinition.Direction, true
	default:
		return "", false
	}
}
