package gitmaterialize

import (
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// MaterializeRequest describes one materialization using existing domain shapes.
type MaterializeRequest struct {
	Task domain.LCATask

	// CacheDir is the root directory for mirrors and checkouts.
	CacheDir domain.HostPath

	// RemoteURL overrides the default https://github.com/{owner}/{name}.git URL.
	RemoteURL string

	ForceRefresh bool
}

// MaterializedRepo is the durable outcome of materializing one task.
type MaterializedRepo struct {
	TaskID domain.MatchID

	RepoOwner string
	RepoName  string
	BaseSHA   string

	RepoURL string

	RootPath domain.HostPath

	CacheKey string

	MaterializedAt time.Time
}

// NewMaterializeRequest builds a request from an LCA task and cache directory.
func NewMaterializeRequest(task domain.LCATask, cacheDir domain.HostPath, forceRefresh bool) MaterializeRequest {
	return MaterializeRequest{
		Task:         task,
		CacheDir:     cacheDir,
		ForceRefresh: forceRefresh,
	}
}

// DefaultGitHubRemoteURL resolves the canonical GitHub HTTPS remote for the task identity.
func DefaultGitHubRemoteURL(id domain.LCATaskIdentity) string {
	owner := strings.TrimSpace(id.RepoOwner)
	name := strings.TrimSpace(id.RepoName)
	return "https://github.com/" + owner + "/" + name + ".git"
}
