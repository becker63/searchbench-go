package score

import "iter"

type MetricPoint struct {
	Name      MetricName
	Value     float64
	Direction Direction
}

type MetricComparison struct {
	Metric    MetricName
	Baseline  float64
	Candidate float64
	Delta     float64
	Direction Direction
	Improved  bool
	Regressed bool
}

// Points returns the full required metric set in stable metric order.
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

func CompareSets(baseline ScoreSet, candidate ScoreSet) []MetricComparison {
	candidateByName := make(map[MetricName]MetricPoint, 5)
	for point := range candidate.Points() {
		candidateByName[point.Name] = point
	}

	out := make([]MetricComparison, 0, 5)
	for base := range baseline.Points() {
		cand, ok := candidateByName[base.Name]
		if !ok {
			continue
		}
		delta := cand.Value - base.Value
		improved := Improved(base.Direction, base.Value, cand.Value)
		out = append(out, MetricComparison{
			Metric:    base.Name,
			Baseline:  base.Value,
			Candidate: cand.Value,
			Delta:     delta,
			Direction: base.Direction,
			Improved:  improved,
			Regressed: base.Value != cand.Value && !improved,
		})
	}
	return out
}

// AverageByMetric returns averages for every required metric name.
//
// Empty slices produce all required metrics with zero values so callers can
// still build a complete comparison shape.
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

func CompareAverages(baseline []ScoredRun, candidate []ScoredRun) []MetricComparison {
	baselineAverages := AverageByMetric(baseline)
	candidateAverages := AverageByMetric(candidate)

	return []MetricComparison{
		newMetricComparison(MetricGoldHop, baselineAverages[MetricGoldHop], candidateAverages[MetricGoldHop]),
		newMetricComparison(MetricIssueHop, baselineAverages[MetricIssueHop], candidateAverages[MetricIssueHop]),
		newMetricComparison(MetricTokenEfficiency, baselineAverages[MetricTokenEfficiency], candidateAverages[MetricTokenEfficiency]),
		newMetricComparison(MetricCost, baselineAverages[MetricCost], candidateAverages[MetricCost]),
		newMetricComparison(MetricComposite, baselineAverages[MetricComposite], candidateAverages[MetricComposite]),
	}
}

func newMetricComparison(name MetricName, baseline float64, candidate float64) MetricComparison {
	direction, _ := DirectionForMetric(name)
	delta := candidate - baseline
	improved := Improved(direction, baseline, candidate)
	return MetricComparison{
		Metric:    name,
		Baseline:  baseline,
		Candidate: candidate,
		Delta:     delta,
		Direction: direction,
		Improved:  improved,
		Regressed: baseline != candidate && !improved,
	}
}
