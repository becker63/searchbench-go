package domain

// ID is a strongly typed identifier.
//
// The type parameter is a phantom tag used to prevent accidentally mixing
// IDs from different domains, such as TaskID and RunID.
type ID[T any] string

func (id ID[T]) String() string {
	return string(id)
}

func (id ID[T]) Empty() bool {
	return id == ""
}

type taskTag struct{}
type runTag struct{}
type systemTag struct{}
type sessionTag struct{}
type policyTag struct{}
type traceTag struct{}
type artifactTag struct{}
type reportTag struct{}
type nodeTag struct{}

type TaskID = ID[taskTag]
type RunID = ID[runTag]
type SystemID = ID[systemTag]
type SessionID = ID[sessionTag]
type PolicyID = ID[policyTag]
type TraceID = ID[traceTag]
type ArtifactID = ID[artifactTag]
type ReportID = ID[reportTag]
type NodeID = ID[nodeTag]
