// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package artifactkind

import (
	"encoding"
	"fmt"
)

type ArtifactKind string

const (
	Policy               ArtifactKind = "policy"
	PolicyProposal       ArtifactKind = "policy_proposal"
	CompletedRoundBundle ArtifactKind = "completed_round_bundle"
)

// String returns the string representation of ArtifactKind
func (rcv ArtifactKind) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(ArtifactKind)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for ArtifactKind.
func (rcv *ArtifactKind) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "policy":
		*rcv = Policy
	case "policy_proposal":
		*rcv = PolicyProposal
	case "completed_round_bundle":
		*rcv = CompletedRoundBundle
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid ArtifactKind`, str)
	}
	return nil
}
