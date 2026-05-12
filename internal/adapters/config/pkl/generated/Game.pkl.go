// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Logical game this round belongs to (drives artifact roots under the game id).
type Game struct {
	// Stable game id (e.g. product bundle namespace).
	Id string `pkl:"id"`

	// Free-form game kind string for downstream classification.
	Kind string `pkl:"kind"`
}
