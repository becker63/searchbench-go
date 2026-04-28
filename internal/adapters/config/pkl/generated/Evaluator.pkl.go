// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

type Evaluator struct {
	Model Model `pkl:"model"`

	Bounds AgentBounds `pkl:"bounds"`

	Retry RetryPolicy `pkl:"retry"`
}
