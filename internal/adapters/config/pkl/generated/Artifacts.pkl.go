// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Durable artifact pointers used when binding evaluation and optimization targets.
type Artifacts struct {
	// Challenger policy file on disk (relative to manifest dir unless otherwise specified by harness).
	ChallengerPolicy *PolicyArtifact `pkl:"challengerPolicy"`

	// Declares the artifact record for the next-challenger proposal output.
	NextChallenger *NextChallengerArtifact `pkl:"nextChallenger"`

	// Completed parent round bundle (for evidence loading in optimization or scored comparisons with a parent).
	ParentRoundBundle *CompletedRoundBundleArtifact `pkl:"parentRoundBundle"`
}
