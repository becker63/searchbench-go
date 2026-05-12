// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Challenger side binding plus artifact reference for the implemented selection policy.
type ChallengerEvaluationBinding struct {
	// System block from `policies` for the challenger role.
	System System `pkl:"system"`

	// Artifact pointers required for challenger execution.
	Uses ChallengerUses `pkl:"uses"`
}
