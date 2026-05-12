// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Hard limits while executing one system on one match.
type Runtime struct {
	// Maximum LLM/agent steps for this system (must be positive).
	MaxSteps int `pkl:"maxSteps"`

	// Wall-clock timeout in seconds for system execution on one match attempt.
	TimeoutSeconds int `pkl:"timeoutSeconds"`
}
