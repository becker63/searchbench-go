// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Evaluation-mode root block: wires agent, policy bindings, scoring objective path, and report outputs.
//
// Must stay consistent with `agents.evaluator` and `policies` (validated in Go).
type Evaluation struct {
	// Snapshot that must equal `agents.evaluator` (same model, bounds, tools, prompts).
	Agent Evaluator `pkl:"agent"`

	// Which manifest `policies.incumbent` entry this evaluation uses for the incumbent role.
	Incumbent EvaluationSystemBinding `pkl:"incumbent"`

	// Which manifest `policies.challenger` entry this evaluation uses, plus required selection policy artifact binding.
	Challenger ChallengerEvaluationBinding `pkl:"challenger"`

	// Path to the Pkl **objective** module that defines scoring over evidence.
	Scoring Scoring `pkl:"scoring"`

	// Report output formats for this round run.
	Report Report `pkl:"report"`
}
