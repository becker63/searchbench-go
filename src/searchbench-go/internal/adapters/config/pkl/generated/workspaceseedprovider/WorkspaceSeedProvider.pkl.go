// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package workspaceseedprovider

import (
	"encoding"
	"fmt"
)

// IC workspace seed provider for optimizable IC-backed systems.
type WorkspaceSeedProvider string

const (
	LocalPath      WorkspaceSeedProvider = "local_path"
	BuckDescriptor WorkspaceSeedProvider = "buck_descriptor"
	Git            WorkspaceSeedProvider = "git"
	Archive        WorkspaceSeedProvider = "archive"
)

// String returns the string representation of WorkspaceSeedProvider
func (rcv WorkspaceSeedProvider) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(WorkspaceSeedProvider)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for WorkspaceSeedProvider.
func (rcv *WorkspaceSeedProvider) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "local_path":
		*rcv = LocalPath
	case "buck_descriptor":
		*rcv = BuckDescriptor
	case "git":
		*rcv = Git
	case "archive":
		*rcv = Archive
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid WorkspaceSeedProvider`, str)
	}
	return nil
}
