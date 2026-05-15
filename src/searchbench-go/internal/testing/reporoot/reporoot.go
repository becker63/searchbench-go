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
	return walkUpToFile(tb, callerDir(tb), "go.mod")
}

func callerDir(tb testing.TB) string {
	tb.Helper()
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		tb.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func walkUpToFile(tb testing.TB, startDir, rel string) string {
	tb.Helper()
	dir := filepath.Clean(startDir)
	for {
		candidate := filepath.Join(dir, rel)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			tb.Fatalf("walk-up: %q not found from %q", rel, startDir)
		}
		dir = parent
	}
}
