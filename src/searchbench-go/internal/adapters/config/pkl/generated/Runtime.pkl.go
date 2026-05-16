// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Hard limits while executing one system on one match.
type Runtime struct {
	MaxSteps int `pkl:"maxSteps"`

	TimeoutSeconds int `pkl:"timeoutSeconds"`

	// Optional IC workspace seed (local path copy into candidate workspace).
	WorkspaceSeed *WorkspaceSeedConfig `pkl:"workspaceSeed"`
}
