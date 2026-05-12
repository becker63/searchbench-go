// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Describes the next challenger artifact slot produced by optimization.
type NextChallengerArtifact struct {
	// Stable artifact id.
	Id string `pkl:"id"`

	// Fixed kind tag.
	Kind string `pkl:"kind"`

	// File name (relative path) for the emitted next challenger policy artifact.
	ArtifactName string `pkl:"artifactName"`

	// Interface the emitted artifact must satisfy.
	Implements Interface `pkl:"implements"`
}
