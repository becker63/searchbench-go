// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Parent round locator inside the optimization graph.
type ParentRound struct {
	// Completed round bundle artifact; must align with `artifacts.parentRoundBundle` (validated in Go).
	Bundle CompletedRoundBundleArtifact `pkl:"bundle"`
}
