// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Pair of executable policy systems compared in this round.
type Policies struct {
	// Baseline system (incumbent policy role).
	Incumbent System `pkl:"incumbent"`

	// Candidate system (challenger policy role).
	Challenger System `pkl:"challenger"`
}
