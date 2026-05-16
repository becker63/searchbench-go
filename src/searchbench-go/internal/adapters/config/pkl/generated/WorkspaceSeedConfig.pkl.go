// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/workspaceseedprovider"

// Declares how an IC optimizable backend seed is obtained for candidate workspaces.
type WorkspaceSeedConfig struct {
	Provider workspaceseedprovider.WorkspaceSeedProvider `pkl:"provider"`

	// Required when provider is local_path (repo-relative or absolute path).
	LocalPath *string `pkl:"localPath"`

	// Required when provider is buck_descriptor (repo-internal Buck label).
	BuckDescriptorTarget *string `pkl:"buckDescriptorTarget"`
}
