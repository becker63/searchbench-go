// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Configuration for the **optimizer** agent (next challenger proposal). Tool allow/deny is validated for shape only until an optimizer tool runtime exists.
type Optimizer struct {
	// Chat model used for optimization / proposal generation.
	Model Model `pkl:"model"`

	// Agent attempt bounds for the optimizer loop.
	Bounds AgentBounds `pkl:"bounds"`

	// Structural tool allow/deny (not enforced against a registry in the current harness).
	Tools AgentToolPolicy `pkl:"tools"`

	// Optional extra instructions for the optimizer prompt appendix (size-capped in Go when validated).
	SystemPrompt *string `pkl:"systemPrompt"`
}
