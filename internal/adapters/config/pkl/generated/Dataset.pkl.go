// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Selects a dataset slice from an upstream bug-localization benchmark.
type Dataset struct {
	Kind string `pkl:"kind"`

	Name string `pkl:"name"`

	Config string `pkl:"config"`

	Split string `pkl:"split"`

	MaxItems *int `pkl:"maxItems"`
}
