// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Optional tracing configuration (feature-flag style).
type Tracing struct {
	// Master switch for emission of traces.
	Enabled bool `pkl:"enabled"`

	// Tracing backend identifier when enabled.
	Provider *string `pkl:"provider"`

	// Optional project or workspace key for the tracer.
	Project *string `pkl:"project"`
}
