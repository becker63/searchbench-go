package materialize

import (
	"path/filepath"
	"strings"
)

// defaultExcludeNames are never copied into IC candidate workspaces.
var defaultExcludeNames = map[string]struct{}{
	".git":               {},
	".venv":              {},
	".pytest_cache":      {},
	".ruff_cache":        {},
	".hypothesis":        {},
	"__pycache__":        {},
	"repomix-output.xml": {},
	"buck-out":           {},
	"result":             {},
}

// ShouldExcludePath reports whether rel (slash-separated, from workspace root) must be skipped.
func ShouldExcludePath(rel string) bool {
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	if rel == "" || rel == "." {
		return false
	}
	parts := strings.Split(rel, "/")
	for _, part := range parts {
		if _, ok := defaultExcludeNames[part]; ok {
			return true
		}
		if strings.HasPrefix(part, "result-") {
			return true
		}
	}
	base := filepath.Base(rel)
	if _, ok := defaultExcludeNames[base]; ok {
		return true
	}
	return false
}
