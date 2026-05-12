// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// A versioned policy artifact implementing a declared interface.
type PolicyArtifact struct {
	// Stable artifact id for cross-manifest references.
	Id string `pkl:"id"`

	// Fixed kind tag for routing and validation.
	Kind string `pkl:"kind"`

	// Repo-relative path to policy source checked in next to the manifest tree.
	Path string `pkl:"path"`

	// Interface contract this policy must implement.
	Implements Interface `pkl:"implements"`
}
