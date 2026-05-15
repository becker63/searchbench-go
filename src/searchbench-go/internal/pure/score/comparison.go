package score

import "iter"

type MetricPoint struct {
	Name      MetricName
	Value     float64
	Direction Direction
}

// MetricComparison compares one metric between incumbent and challenger.
type MetricComparison struct {
	Metric     MetricName
	Incumbent  float64
	Challenger float64
	Delta      float64
	Direction  Direction
	Improved   bool
	Regressed  bool
}

// Points returns the full required metric set in stable metric order.
//
// Stable ordering keeps metric iteration deterministic for comparison and
// reporting code.
func (s ScoreSet) Points() iter.Seq[MetricPoint] {
	return func(yield func(MetricPoint) bool) {
		if !yield(MetricPoint{
			Name:      MetricGoldHop,
			Value:     float64(s.GoldHop.Value),
			Direction: GoldHopDefinition.Direction,
		}) {
			return
		}
		if !yield(MetricPoint{
			Name:      MetricIssueHop,
			Value:     float64(s.IssueHop.Value),
			Direction: IssueHopDefinition.Direction,
		}) {
			return
		}
		if !yield(MetricPoint{
			Name:      MetricTokenEfficiency,
			Value:     float64(s.TokenEfficiency.Value),
			Direction: TokenEfficiencyDefinition.Direction,
		}) {
			return
		}
		if !yield(MetricPoint{
			Name:      MetricCost,
			Value:     float64(s.Cost.Value),
			Direction: CostDefinition.Direction,
		}) {
			return
		}
		yield(MetricPoint{
			Name:      MetricComposite,
			Value:     float64(s.Composite.Value),
			Direction: CompositeDefinition.Direction,
		})
	}
}

// CompareSets compares two complete score sets metric-by-metric.
func CompareSets(incumbent ScoreSet, challenger ScoreSet) []MetricComparison {
	challengerByName := make(map[MetricName]MetricPoint, 5)
	for point := range challenger.Points() {
		challengerByName[point.Name] = point
	}

	out := make([]MetricComparison, 0, 5)
	for inc := range incumbent.Points() {
		chal, ok := challengerByName[inc.Name]
		if !ok {
			continue
		}
		delta := chal.Value - inc.Value
		improved := Improved(inc.Direction, inc.Value, chal.Value)
		out = append(out, MetricComparison{
			Metric:     inc.Name,
			Incumbent:  inc.Value,
			Challenger: chal.Value,
			Delta:      delta,
			Direction:  inc.Direction,
			Improved:   improved,
			Regressed:  inc.Value != chal.Value && !improved,
		})
	}
	return out
}

// AverageByMetric returns averages for every required metric name.
//
// Empty slices produce all required metrics with zero values so callers can
// still build a complete comparison shape. This is a comparison convenience,
// not a promotion policy; future release rules should treat missing run sets
// explicitly.
func AverageByMetric(runs []ScoredRun) map[MetricName]float64 {
	totals := map[MetricName]float64{
		MetricGoldHop:         0,
		MetricIssueHop:        0,
		MetricTokenEfficiency: 0,
		MetricCost:            0,
		MetricComposite:       0,
	}
	if len(runs) == 0 {
		return totals
	}

	for _, run := range runs {
		for point := range run.Scores.Points() {
			totals[point.Name] += point.Value
		}
	}
	for name, total := range totals {
		totals[name] = total / float64(len(runs))
	}
	return totals
}

// CompareAverages compares incumbent and challenger run sets using per-metric
// averages.
func CompareAverages(incumbent []ScoredRun, challenger []ScoredRun) []MetricComparison {
	incumbentAverages := AverageByMetric(incumbent)
	challengerAverages := AverageByMetric(challenger)

	return []MetricComparison{
		newMetricComparison(MetricGoldHop, incumbentAverages[MetricGoldHop], challengerAverages[MetricGoldHop]),
		newMetricComparison(MetricIssueHop, incumbentAverages[MetricIssueHop], challengerAverages[MetricIssueHop]),
		newMetricComparison(MetricTokenEfficiency, incumbentAverages[MetricTokenEfficiency], challengerAverages[MetricTokenEfficiency]),
		newMetricComparison(MetricCost, incumbentAverages[MetricCost], challengerAverages[MetricCost]),
		newMetricComparison(MetricComposite, incumbentAverages[MetricComposite], challengerAverages[MetricComposite]),
	}
}

func newMetricComparison(name MetricName, incumbent float64, challenger float64) MetricComparison {
	direction, _ := DirectionForMetric(name)
	delta := challenger - incumbent
	improved := Improved(direction, incumbent, challenger)
	return MetricComparison{
		Metric:     name,
		Incumbent:  incumbent,
		Challenger: challenger,
		Delta:      delta,
		Direction:  direction,
		Improved:   improved,
		Regressed:  incumbent != challenger && !improved,
	}
}
