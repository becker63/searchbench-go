package score

// MetricName identifies a required Searchbench metric.
type MetricName string

const (
	MetricGoldHop         MetricName = "gold_hop"
	MetricIssueHop        MetricName = "issue_hop"
	MetricTokenEfficiency MetricName = "token_efficiency"
	MetricCost            MetricName = "cost"
	MetricComposite       MetricName = "composite"
)

// HopDistance is a graph distance measured in hops.
//
// Lower is usually better.
type HopDistance int

// EfficiencyScore represents token/cost efficiency.
//
// The exact formula can evolve, but the type keeps it distinct from hop distance.
type EfficiencyScore float64

// CostScore represents normalized cost performance.
type CostScore float64

// CompositeScore is the final reduced score used for comparison/promotion.
type CompositeScore float64

// EvidenceRef points to evidence used to compute or explain a metric.
type EvidenceRef struct {
	Kind string `json:"kind"`
	Ref  string `json:"ref"`
}

// Metric is a typed metric value.
//
// The type parameter prevents mixing incompatible metric value kinds, such as
// using a token-efficiency value where a hop distance is expected.
type Metric[T any] struct {
	Name        MetricName    `json:"name"`
	Value       T             `json:"value"`
	Explanation string        `json:"explanation,omitempty"`
	Evidence    []EvidenceRef `json:"evidence,omitempty"`
}
