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

type ReportArtifactRef = ArtifactRef[reportArtifactTag]
type TraceArtifactRef = ArtifactRef[traceArtifactTag]
type GraphArtifactRef = ArtifactRef[graphArtifactTag]
type PolicyArtifactRef = ArtifactRef[policyArtifactTag]
type DatasetArtifactRef = ArtifactRef[datasetArtifactTag]
