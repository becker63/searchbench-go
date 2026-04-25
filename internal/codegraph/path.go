package codegraph

// Path is a graph path between two nodes.
type Path struct {
	Nodes []NodeID `json:"nodes"`
	Edges []Edge   `json:"edges"`
	Cost  float64  `json:"cost"`
}

func (p Path) Found() bool {
	return len(p.Nodes) > 0
}

func (p Path) Hops() int {
	if len(p.Nodes) == 0 {
		return 0
	}
	return len(p.Nodes) - 1
}
