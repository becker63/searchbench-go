// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Configuration for the SearchBench **evaluator** agent (compares predictions on matches).
type Evaluator struct {
	// Primary chat model used by the evaluator runtime.
	Model Model `pkl:"model"`

	// Loop and wall-clock bounds for evaluator attempts.
	Bounds AgentBounds `pkl:"bounds"`

	// Tool allow/deny policy surfaced from the manifest; intersected with the evaluator runtime registry in Go.
	Tools AgentToolPolicy `pkl:"tools"`

	// Optional extra system-facing instructions appended to the stable evaluator prompt XML (trimmed; size-capped in Go).
	SystemPrompt *string `pkl:"systemPrompt"`

	// Retry behavior across evaluator attempts (distinct from model “turns” inside one attempt).
	Retry RetryPolicy `pkl:"retry"`

	// Optional telemetry hooks (reserved / product-specific).
	Tracing *Tracing `pkl:"tracing"`
}
