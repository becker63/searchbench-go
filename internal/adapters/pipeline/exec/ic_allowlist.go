package execpipeline

import (
	"path/filepath"
	"slices"
	"strings"
)

// ICOptimizerAllowlist permits the argv shapes used by Iterative Context optimizer
// validation. Dynamic path/id/symbol segments are checked against the staged workspace.
type ICOptimizerAllowlist struct {
	UvBinary         string
	WorkspaceRootAbs string
	PolicyFileAbs    string
	PolicyID         string
	Symbol           string
}

// Allows implements CommandAllowlist for IC optimizer pipelines.
func (a ICOptimizerAllowlist) Allows(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	if filepath.Base(argv[0]) != "uv" {
		return false
	}
	uvBin := strings.TrimSpace(a.UvBinary)
	if uvBin != "" && argv[0] != uvBin {
		return false
	}
	switch {
	case slices.Equal(argv, []string{a.UvBinary, "run", "basedpyright"}):
		return true
	case slices.Equal(argv, []string{a.UvBinary, "run", "ruff", "check"}):
		return true
	case slices.Equal(argv, []string{a.UvBinary, "run", "pytest"}):
		return true
	default:
		return a.allowsValidatePolicy(argv)
	}
}

func (a ICOptimizerAllowlist) allowsValidatePolicy(argv []string) bool {
	want := []string{
		a.UvBinary, "run", "python", "-m", "iterative_context.validate_policy",
		"--policy-path", a.PolicyFileAbs,
		"--policy-id", a.PolicyID,
		"--symbol", a.Symbol,
		"--json",
	}
	if !slices.Equal(argv, want) {
		return false
	}
	wRoot, err := filepath.EvalSymlinks(filepath.Clean(a.WorkspaceRootAbs))
	if err != nil {
		wRoot = filepath.Clean(a.WorkspaceRootAbs)
	}
	pFile, err := filepath.EvalSymlinks(filepath.Clean(a.PolicyFileAbs))
	if err != nil {
		pFile = filepath.Clean(a.PolicyFileAbs)
	}
	rel, err := filepath.Rel(wRoot, pFile)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
