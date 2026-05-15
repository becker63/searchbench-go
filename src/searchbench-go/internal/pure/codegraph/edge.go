package codegraph

// EdgeKind identifies the semantic relationship between two code graph nodes.
type EdgeKind string

const (
	EdgeContains   EdgeKind = "contains"
	EdgeDefines    EdgeKind = "defines"
	EdgeCalls      EdgeKind = "calls"
	EdgeImports    EdgeKind = "imports"
	EdgeReferences EdgeKind = "references"
)

// Edge is a directed semantic edge in the code graph.
type Edge struct {
	From   NodeID   `json:"from"`
	To     NodeID   `json:"to"`
	Kind   EdgeKind `json:"kind"`
	Weight float64  `json:"weight,omitempty"`
}

// NewEdge constructs an edge with the default weight.
func NewEdge(from, to NodeID, kind EdgeKind) Edge {
	return Edge{
		From:   from,
		To:     to,
		Kind:   kind,
		Weight: 1,
	}
}

// NewWeightedEdge constructs an edge with an explicit weight.
func NewWeightedEdge(from, to NodeID, kind EdgeKind, weight float64) Edge {
	if weight == 0 {
		weight = 1
	}
	return Edge{
		From:   from,
		To:     to,
		Kind:   kind,
		Weight: weight,
	}
}
