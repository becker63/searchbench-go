// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Locates the prompt bundle material for a system (version pins reproducibility).
type PromptBundle struct {
	// Bundle name known to the product (e.g. `default`).
	Name string `pkl:"name"`

	// Optional version or channel label for the bundle.
	Version *string `pkl:"version"`
}
