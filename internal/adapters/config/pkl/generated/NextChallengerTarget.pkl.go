// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Declares input policy to improve and the next challenger output artifact record.
type NextChallengerTarget struct {
	// Current challenger policy artifact that optimization reads as input.
	Input PolicyArtifact `pkl:"input"`

	// Artifact record describing the file name and interface for the proposed next policy.
	Output NextChallengerArtifact `pkl:"output"`
}
