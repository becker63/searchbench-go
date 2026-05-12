// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Configuration for the optimizer agent used to materialize a challenger.
type Optimizer struct {
	Model Model `pkl:"model"`

	Bounds AgentBounds `pkl:"bounds"`

	Tools AgentToolPolicy `pkl:"tools"`

	SystemPrompt *string `pkl:"systemPrompt"`
}
