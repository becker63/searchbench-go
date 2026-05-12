// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Shared safety bounds for agent loops (evaluator, optimizer).
type AgentBounds struct {
	// Maximum model turns per attempt in the agent harness.
	MaxModelTurns int `pkl:"maxModelTurns"`

	// Maximum tool invocations per attempt (non-negative; `0` may disable tools depending on runtime).
	MaxToolCalls int `pkl:"maxToolCalls"`

	// Wall-clock ceiling in seconds for a single agent attempt sequence.
	TimeoutSeconds int `pkl:"timeoutSeconds"`
}
