package gitmaterialize

import (
	"path/filepath"
	"strings"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

// CacheKey returns the deterministic cache segment for one repo@sha (owner,
// name, base SHA normalized like domain identity tokens).
func CacheKey(owner, name, baseSHA string) string {
	o := strings.ToLower(strings.TrimSpace(owner))
	n := strings.ToLower(strings.TrimSpace(name))
	s := strings.ToLower(strings.TrimSpace(baseSHA))
	return filepath.Join(o, n, s)
}

// CacheKeyFromTask delegates to identity fields on the LCA task.
func CacheKeyFromTask(t domain.LCATask) string {
	id := t.Identity
	return CacheKey(id.RepoOwner, id.RepoName, id.BaseSHA)
}
