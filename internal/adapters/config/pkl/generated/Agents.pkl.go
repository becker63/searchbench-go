// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Agent definitions for roles that participate in round resolution.
type Agents struct {
	// Evaluator agent configuration (required when `mode` is `evaluation` and for `evaluation.*` blocks).
	Evaluator *Evaluator `pkl:"evaluator"`

	// Optimizer agent configuration (required when `mode` is `optimization` and for `optimization.*` blocks).
	Optimizer *Optimizer `pkl:"optimizer"`
}
