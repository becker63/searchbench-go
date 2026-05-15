package codegraph

// Builder is the mutable construction-time interface used while creating a
// graph.
//
// Tree-sitter ingestion should receive Builder. Scoring should depend on Graph
// instead so it does not rely on a concrete mutable store.
type Builder interface {
	AddNode(node Node) error
	AddEdge(edge Edge) error
	Build() (Graph, error)
}
