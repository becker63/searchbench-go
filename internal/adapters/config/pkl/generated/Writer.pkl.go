// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

type Writer struct {
	Enabled bool `pkl:"enabled"`

	Model Model `pkl:"model"`

	MaxAttempts int `pkl:"maxAttempts"`

	Pipeline *PipelineProfile `pkl:"pipeline"`
}
