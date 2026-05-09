// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

type Optimization struct {
	Agent Optimizer `pkl:"agent"`

	ParentRound ParentRound `pkl:"parentRound"`

	Target OptimizationTarget `pkl:"target"`

	Evidence OptimizationEvidence `pkl:"evidence"`
}
