package score

import "fmt"

// ScoreSet is the complete required metric set for one executed run.
//
// If a ScoreSet exists, every required metric exists. If any metric cannot be
// computed, scoring should fail instead of constructing a partial ScoreSet.
type ScoreSet struct {
	GoldHop         Metric[HopDistance]     `json:"gold_hop"`
	IssueHop        Metric[HopDistance]     `json:"issue_hop"`
	TokenEfficiency Metric[EfficiencyScore] `json:"token_efficiency"`
	Cost            Metric[CostScore]       `json:"cost"`
	Composite       Metric[CompositeScore]  `json:"composite"`
}

// NewScoreSet constructs a complete required score set and validates metric
// names.
func NewScoreSet(
	goldHop Metric[HopDistance],
	issueHop Metric[HopDistance],
	tokenEfficiency Metric[EfficiencyScore],
	cost Metric[CostScore],
	composite Metric[CompositeScore],
) (ScoreSet, error) {
	s := ScoreSet{
		GoldHop:         goldHop,
		IssueHop:        issueHop,
		TokenEfficiency: tokenEfficiency,
		Cost:            cost,
		Composite:       composite,
	}
	if err := s.Validate(); err != nil {
		return ScoreSet{}, err
	}
	return s, nil
}

// Validate checks that the score set contains the required metrics in their
// canonical slots.
//
// ScoreSet intentionally does not use per-metric Available flags. Missing
// required metrics should be represented as scoring failure, not as a partial
// score set.
func (s ScoreSet) Validate() error {
	if s.GoldHop.Name != MetricGoldHop {
		return fmt.Errorf("gold hop metric must be named %q", MetricGoldHop)
	}
	if s.IssueHop.Name != MetricIssueHop {
		return fmt.Errorf("issue hop metric must be named %q", MetricIssueHop)
	}
	if s.TokenEfficiency.Name != MetricTokenEfficiency {
		return fmt.Errorf("token efficiency metric must be named %q", MetricTokenEfficiency)
	}
	if s.Cost.Name != MetricCost {
		return fmt.Errorf("cost metric must be named %q", MetricCost)
	}
	if s.Composite.Name != MetricComposite {
		return fmt.Errorf("composite metric must be named %q", MetricComposite)
	}
	return nil
}

// Metrics returns the score set as a flat list for report-facing inspection.
//
// This is intentionally read-only-ish: callers can inspect metrics uniformly,
// but the canonical complete representation is still the struct above.
func (s ScoreSet) Metrics() []AnyMetric {
	return []AnyMetric{
		{
			Name:        s.GoldHop.Name,
			Value:       float64(s.GoldHop.Value),
			Explanation: s.GoldHop.Explanation,
			Evidence:    s.GoldHop.Evidence,
		},
		{
			Name:        s.IssueHop.Name,
			Value:       float64(s.IssueHop.Value),
			Explanation: s.IssueHop.Explanation,
			Evidence:    s.IssueHop.Evidence,
		},
		{
			Name:        s.TokenEfficiency.Name,
			Value:       float64(s.TokenEfficiency.Value),
			Explanation: s.TokenEfficiency.Explanation,
			Evidence:    s.TokenEfficiency.Evidence,
		},
		{
			Name:        s.Cost.Name,
			Value:       float64(s.Cost.Value),
			Explanation: s.Cost.Explanation,
			Evidence:    s.Cost.Evidence,
		},
		{
			Name:        s.Composite.Name,
			Value:       float64(s.Composite.Value),
			Explanation: s.Composite.Explanation,
			Evidence:    s.Composite.Evidence,
		},
	}
}

// AnyMetric is a normalized report-facing metric value.
//
// Keep this as an output/read model. Do not use it as the authoritative scoring
// model, because that would reintroduce partial/dynamic metric behavior.
type AnyMetric struct {
	Name        MetricName    `json:"name"`
	Value       float64       `json:"value"`
	Explanation string        `json:"explanation,omitempty"`
	Evidence    []EvidenceRef `json:"evidence,omitempty"`
}
