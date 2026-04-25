package codegraph

import (
	"fmt"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/hmdsefi/gograph"
)

var _ Builder = (*Store)(nil)
var _ Graph = (*Store)(nil)

// Store is the current in-memory graph implementation.
//
// Store currently implements both Builder and Graph so callers can construct a
// graph and then query it through one type. Its gograph dependency is an
// internal detail hidden behind this package.
type Store struct {
	graph gograph.Graph[NodeID]

	nodes map[NodeID]Node
	edges []Edge

	nodesByFile     map[domain.RepoRelPath][]NodeID
	functionsByName map[string][]NodeID
}

// NewStore constructs an empty graph store.
func NewStore() *Store {
	return &Store{
		graph:           gograph.New[NodeID](gograph.Directed()),
		nodes:           make(map[NodeID]Node),
		edges:           make([]Edge, 0),
		nodesByFile:     make(map[domain.RepoRelPath][]NodeID),
		functionsByName: make(map[string][]NodeID),
	}
}

// AddNode inserts one node into the mutable store.
func (s *Store) AddNode(node Node) error {
	if node.ID == "" {
		return fmt.Errorf("codegraph: node id is required")
	}

	if _, exists := s.nodes[node.ID]; exists {
		return fmt.Errorf("codegraph: duplicate node id %q", node.ID)
	}

	s.nodes[node.ID] = node
	s.graph.AddVertexByLabel(node.ID)

	if file, ok := node.RepoFile(); ok {
		s.nodesByFile[file] = append(s.nodesByFile[file], node.ID)
	}

	if node.Kind == NodeFunction && node.Function != nil {
		s.functionsByName[node.Function.Name] = append(s.functionsByName[node.Function.Name], node.ID)
	}

	return nil
}

// AddEdge inserts one edge into the mutable store.
func (s *Store) AddEdge(edge Edge) error {
	if edge.From == "" || edge.To == "" {
		return fmt.Errorf("codegraph: edge endpoints are required")
	}
	if _, ok := s.nodes[edge.From]; !ok {
		return fmt.Errorf("codegraph: missing from node %q", edge.From)
	}
	if _, ok := s.nodes[edge.To]; !ok {
		return fmt.Errorf("codegraph: missing to node %q", edge.To)
	}
	if edge.Weight == 0 {
		edge.Weight = 1
	}

	from := s.graph.GetVertexByID(edge.From)
	to := s.graph.GetVertexByID(edge.To)
	if from == nil || to == nil {
		return fmt.Errorf("codegraph: missing graph vertex for edge %q -> %q", edge.From, edge.To)
	}

	_, err := s.graph.AddEdge(from, to, gograph.WithEdgeWeight(edge.Weight))
	if err != nil {
		return err
	}

	s.edges = append(s.edges, edge)
	return nil
}

// Build returns the store as a read/query-time Graph.
func (s *Store) Build() (Graph, error) {
	return s, nil
}

// Node looks up one node by ID.
func (s *Store) Node(id NodeID) (Node, bool) {
	node, ok := s.nodes[id]
	return node, ok
}

// Nodes returns all nodes currently stored.
func (s *Store) Nodes() []Node {
	out := make([]Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		out = append(out, node)
	}
	return out
}

// Edges returns all edges currently stored.
func (s *Store) Edges() []Edge {
	out := make([]Edge, len(s.edges))
	copy(out, s.edges)
	return out
}

// Neighbors returns outgoing neighbors that pass the filter.
func (s *Store) Neighbors(id NodeID, filter EdgeFilter) []NodeID {
	out := make([]NodeID, 0)
	for _, edge := range s.edges {
		if edge.From == id && filter.Allows(edge.Kind) {
			out = append(out, edge.To)
		}
	}
	return out
}

// ShortestPath computes a filtered shortest path between two nodes.
func (s *Store) ShortestPath(from NodeID, to NodeID, filter EdgeFilter) (Path, bool) {
	if _, ok := s.nodes[from]; !ok {
		return Path{}, false
	}
	if _, ok := s.nodes[to]; !ok {
		return Path{}, false
	}

	return s.filteredShortestPath(from, to, filter)
}

// NodesByFile returns all node IDs associated with a repo-relative file.
func (s *Store) NodesByFile(file domain.RepoRelPath) []NodeID {
	ids := s.nodesByFile[file]
	out := make([]NodeID, len(ids))
	copy(out, ids)
	return out
}

// FunctionsByName returns all function node IDs with the given name.
func (s *Store) FunctionsByName(name string) []NodeID {
	ids := s.functionsByName[name]
	out := make([]NodeID, len(ids))
	copy(out, ids)
	return out
}

func (s *Store) filteredShortestPath(from NodeID, to NodeID, filter EdgeFilter) (Path, bool) {
	type item struct {
		ID   NodeID
		Path []NodeID
	}

	seen := map[NodeID]bool{from: true}
	queue := []item{{ID: from, Path: []NodeID{from}}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		if cur.ID == to {
			return Path{
				Nodes: cur.Path,
				Edges: s.pathEdges(cur.Path),
				Cost:  float64(max(0, len(cur.Path)-1)),
			}, true
		}

		for _, next := range s.Neighbors(cur.ID, filter) {
			if seen[next] {
				continue
			}

			seen[next] = true

			nextPath := make([]NodeID, 0, len(cur.Path)+1)
			nextPath = append(nextPath, cur.Path...)
			nextPath = append(nextPath, next)

			queue = append(queue, item{
				ID:   next,
				Path: nextPath,
			})
		}
	}

	return Path{}, false
}

func (s *Store) pathEdges(nodes []NodeID) []Edge {
	if len(nodes) < 2 {
		return nil
	}

	out := make([]Edge, 0, len(nodes)-1)
	for i := 0; i < len(nodes)-1; i++ {
		from := nodes[i]
		to := nodes[i+1]
		for _, edge := range s.edges {
			if edge.From == from && edge.To == to {
				out = append(out, edge)
				break
			}
		}
	}

	return out
}
