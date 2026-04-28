package domain

// ID is a strongly typed identifier.
//
// The type parameter is a phantom tag used to prevent accidentally mixing
// IDs from different domains, such as TaskID and RunID.
type ID[T any] string

// String returns the identifier as its raw string value.
func (id ID[T]) String() string {
	return string(id)
}

// Empty reports whether the identifier has no value.
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

// TaskID identifies one benchmark task.
type TaskID = ID[taskTag]

// RunID identifies one planned/executed comparison run.
type RunID = ID[runTag]

// SystemID identifies one executable system configuration.
type SystemID = ID[systemTag]

// SessionID identifies one backend session.
type SessionID = ID[sessionTag]

// PolicyID identifies one policy artifact.
type PolicyID = ID[policyTag]

// TraceID identifies one trace or observability artifact.
type TraceID = ID[traceTag]

// ArtifactID identifies one emitted or consumed artifact.
type ArtifactID = ID[artifactTag]

// ReportID identifies one candidate report.
type ReportID = ID[reportTag]

// NodeID identifies one domain-level node reference.
type NodeID = ID[nodeTag]
