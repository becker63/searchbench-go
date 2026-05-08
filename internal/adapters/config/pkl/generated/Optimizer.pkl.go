// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

type Optimizer struct {
	Model Model `pkl:"model"`

	Bounds AgentBounds `pkl:"bounds"`

	Tools AgentToolPolicy `pkl:"tools"`

	SystemPrompt *string `pkl:"systemPrompt"`
}
