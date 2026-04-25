package codegraph

// Builder is the mutation interface used while constructing a graph.
//
// Tree-sitter ingestion should receive Builder. Scoring should receive Graph.
type Builder interface {
	AddNode(node Node) error
	AddEdge(edge Edge) error
	Build() (Graph, error)
}
