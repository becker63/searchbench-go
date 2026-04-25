package score

type Direction string

const (
	LowerIsBetter  Direction = "lower_is_better"
	HigherIsBetter Direction = "higher_is_better"
)

type MetricDefinition[T any] struct {
	Name      MetricName
	Direction Direction
}

var GoldHopDefinition = MetricDefinition[HopDistance]{
	Name:      MetricGoldHop,
	Direction: LowerIsBetter,
}

var IssueHopDefinition = MetricDefinition[HopDistance]{
	Name:      MetricIssueHop,
	Direction: LowerIsBetter,
}

var TokenEfficiencyDefinition = MetricDefinition[EfficiencyScore]{
	Name:      MetricTokenEfficiency,
	Direction: HigherIsBetter,
}

var CostDefinition = MetricDefinition[CostScore]{
	Name:      MetricCost,
	Direction: LowerIsBetter,
}

var CompositeDefinition = MetricDefinition[CompositeScore]{
	Name:      MetricComposite,
	Direction: HigherIsBetter,
}

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
