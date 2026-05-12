// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Configuration for the SearchBench evaluator agent.
type Evaluator struct {
	Model Model `pkl:"model"`

	Bounds AgentBounds `pkl:"bounds"`

	Tools AgentToolPolicy `pkl:"tools"`

	SystemPrompt *string `pkl:"systemPrompt"`

	Retry RetryPolicy `pkl:"retry"`
}
