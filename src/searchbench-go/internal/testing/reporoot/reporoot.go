// Package reporoot resolves repository layout paths for tests.
//
// The Go module lives under src/searchbench-go; Pkl schemas and round configs stay at the monorepo root.
package reporoot

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const schemaMarker = "configs/schema/SearchBenchRound.pkl"

// MonorepoRoot is the git checkout root (directory containing configs/schema/SearchBenchRound.pkl).
func MonorepoRoot(tb testing.TB) string {
	tb.Helper()
	return walkUpToFile(tb, callerDir(tb), schemaMarker)
}

// GoModuleRoot is the directory containing go.mod for the SearchBench Go module.
func GoModuleRoot(tb testing.TB) string {
	tb.Helper()
	if root, ok := walkUpToFileMaybe(callerDir(tb), "go.mod"); ok {
		return root
	}
	return filepath.Join(walkUpToFile(tb, callerDir(tb), filepath.Join("src", "searchbench-go", "go.mod")), "src", "searchbench-go")
}

func callerDir(tb testing.TB) string {
	tb.Helper()
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		tb.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func walkUpToFile(tb testing.TB, startDir, rel string) string {
	tb.Helper()
	if root, ok := walkUpToFileMaybe(startDir, rel); ok {
		return root
	}
	tb.Fatalf("walk-up: %q not found from %q", rel, startDir)
	return ""
}

func walkUpToFileMaybe(startDir, rel string) (string, bool) {
	dir := filepath.Clean(startDir)
	for {
		candidate := filepath.Join(dir, rel)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
