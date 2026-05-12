// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Selects a dataset slice from an upstream bug-localization benchmark.
type Dataset struct {
	// Dataset family; constrains how the harness loads instances.
	Kind string `pkl:"kind"`

	// Upstream dataset repository or product identifier.
	Name string `pkl:"name"`

	// Profile/config key interpreted by the dataset adapter (e.g. language stack).
	Config string `pkl:"config"`

	// Split name within the dataset (e.g. train/dev).
	Split string `pkl:"split"`

	// Optional cap on the number of matches to draw; omit for the adapter default.
	MaxItems *int `pkl:"maxItems"`
}
