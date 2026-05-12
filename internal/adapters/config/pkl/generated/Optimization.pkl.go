// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Optimization-mode root block: proposes a next challenger policy using evidence from a completed parent round.
type Optimization struct {
	// Must mirror `agents.optimizer` (same model, bounds, tools, prompts).
	Agent Optimizer `pkl:"agent"`

	// Which completed bundle supplies parent-round context.
	ParentRound ParentRound `pkl:"parentRound"`

	// Input policy artifact and output next-challenger artifact slots for the proposal.
	Target NextChallengerTarget `pkl:"target"`

	// Which evidence streams to include / forbid when prompting the optimizer.
	Evidence NextChallengerEvidence `pkl:"evidence"`
}
