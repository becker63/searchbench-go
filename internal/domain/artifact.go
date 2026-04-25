package domain

// ArtifactRef is a typed reference to an emitted or consumed artifact.
//
// The type parameter is a phantom tag used to prevent mixing artifact kinds.
type ArtifactRef[T any] struct {
	ID     ArtifactID `json:"id"`
	Path   HostPath   `json:"path,omitempty"`
	SHA256 string     `json:"sha256,omitempty"`
}

type reportArtifactTag struct{}
type traceArtifactTag struct{}
type graphArtifactTag struct{}
type policyArtifactTag struct{}
type datasetArtifactTag struct{}

// ReportArtifactRef is a typed reference to a report artifact.
type ReportArtifactRef = ArtifactRef[reportArtifactTag]

// TraceArtifactRef is a typed reference to a trace artifact.
type TraceArtifactRef = ArtifactRef[traceArtifactTag]

// GraphArtifactRef is a typed reference to a graph artifact.
type GraphArtifactRef = ArtifactRef[graphArtifactTag]

// PolicyArtifactRef is a typed reference to a policy artifact.
type PolicyArtifactRef = ArtifactRef[policyArtifactTag]

// DatasetArtifactRef is a typed reference to a dataset artifact.
type DatasetArtifactRef = ArtifactRef[datasetArtifactTag]
