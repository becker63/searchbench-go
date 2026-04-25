package codegraph

import "github.com/becker63/searchbench-go/internal/domain"

// EdgeFilter limits graph traversal to a subset of edge kinds.
//
// An empty filter means all edge kinds are allowed.
type EdgeFilter struct {
	Kinds []EdgeKind `json:"kinds,omitempty"`
}

func (f EdgeFilter) Allows(kind EdgeKind) bool {
	if len(f.Kinds) == 0 {
		return true
	}
	for _, allowed := range f.Kinds {
		if allowed == kind {
			return true
		}
	}
	return false
}

// Graph is the read-only Searchbench code graph interface.
//
// Scoring depends on this interface, not on gograph directly.
type Graph interface {
	Node(id NodeID) (Node, bool)
	Nodes() []Node
	Edges() []Edge

	Neighbors(id NodeID, filter EdgeFilter) []NodeID
	ShortestPath(from NodeID, to NodeID, filter EdgeFilter) (Path, bool)

	NodesByFile(path domain.RepoRelPath) []NodeID
	FunctionsByName(name string) []NodeID
}
