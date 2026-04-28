package domain

// RepoName identifies a repository in benchmark/task space.
//
// Examples:
//   - "accenture/ampligraph"
//   - "django/django"
//   - "JetBrains-Research/lca-bug-localization:py:dev:..."
type RepoName string

// RepoSHA identifies the exact commit under evaluation.
type RepoSHA string

// RepoRelPath is a path relative to the repository root.
//
// Use this for files returned by agents, gold labels, graph nodes, etc.
type RepoRelPath string

// HostPath is a local filesystem path on the machine running Searchbench.
//
// Keep this distinct from RepoRelPath so agents/tests do not casually mix
// benchmark-relative paths with local paths.
type HostPath string

// RepoSnapshot identifies the exact code universe for a task.
//
// The same RepoSnapshot should be shared by baseline and candidate runs.
type RepoSnapshot struct {
	Name RepoName `json:"name"`
	SHA  RepoSHA  `json:"sha"`
	Path HostPath `json:"path"`
}
