// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Shared safety bounds for agent loops.
type AgentBounds struct {
	MaxModelTurns int `pkl:"maxModelTurns"`

	MaxToolCalls int `pkl:"maxToolCalls"`

	TimeoutSeconds int `pkl:"timeoutSeconds"`
}
