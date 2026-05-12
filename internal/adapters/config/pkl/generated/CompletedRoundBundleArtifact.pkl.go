// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Pointer to an on-disk completed round bundle directory (evidence, reports, metadata).
type CompletedRoundBundleArtifact struct {
	// Stable id referencing this completed bundle in other blocks.
	Id string `pkl:"id"`

	// Fixed kind tag.
	Kind string `pkl:"kind"`

	// Relative path from the manifest to the bundle root (normalized by the harness).
	Path string `pkl:"path"`
}
