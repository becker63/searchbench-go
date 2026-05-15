// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Author-facing round surface: either define a round from scratch or continue
// from an explicit completed bundle path.
type RoundManifest struct {
	Id string `pkl:"id"`

	Continues *string `pkl:"continues"`

	Incumbent *RoundPolicy `pkl:"incumbent"`

	Challenger RoundChallenger `pkl:"challenger"`

	Matches *Dataset `pkl:"matches"`

	Evaluator *Evaluator `pkl:"evaluator"`

	Scoring *Scoring `pkl:"scoring"`

	Report Report `pkl:"report"`
}
