package score

// MetricName identifies one required Searchbench metric.
type MetricName string

const (
	// MetricGoldHop is lower-is-better distance to the gold target.
	MetricGoldHop MetricName = "gold_hop"
	// MetricIssueHop is lower-is-better distance to the issue anchor.
	MetricIssueHop MetricName = "issue_hop"
	// MetricTokenEfficiency is higher-is-better token efficiency.
	MetricTokenEfficiency MetricName = "token_efficiency"
	// MetricCost is lower-is-better cost performance.
	MetricCost MetricName = "cost"
	// MetricComposite is the higher-is-better reduced comparison score.
	MetricComposite MetricName = "composite"
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
