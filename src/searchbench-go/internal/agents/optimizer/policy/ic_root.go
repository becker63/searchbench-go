package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveIterativeContextRoot finds the iterative-context submodule checkout.
// SEARCHBENCH_ITERATIVE_CONTEXT_ROOT overrides discovery.
func ResolveIterativeContextRoot() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("SEARCHBENCH_ITERATIVE_CONTEXT_ROOT")); dir != "" {
		return filepath.Abs(dir)
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	for d := wd; ; d = filepath.Dir(d) {
		legacy := filepath.Join(d, "iterative-context", "pyproject.toml")
		if st, statErr := os.Stat(legacy); statErr == nil && !st.IsDir() {
			return filepath.Abs(filepath.Join(d, "iterative-context"))
		}
		nested := filepath.Join(d, "src", "iterative-context", "pyproject.toml")
		if st, statErr := os.Stat(nested); statErr == nil && !st.IsDir() {
			return filepath.Abs(filepath.Join(d, "src", "iterative-context"))
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
	}
	return "", fmt.Errorf("iterative-context submodule not found from %q (set SEARCHBENCH_ITERATIVE_CONTEXT_ROOT)", wd)
}
