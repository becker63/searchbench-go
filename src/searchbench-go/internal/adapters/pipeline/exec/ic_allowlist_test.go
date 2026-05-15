package execpipeline

import (
	"path/filepath"
	"testing"
)

func TestICOptimizerAllowlistAcceptsValidatePolicyArgv(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	policyPath := filepath.Join(tmp, "next_policy.py")
	ws := tmp
	id := "next-challenger-round-001"
	sym := "score_fn"

	al := ICOptimizerAllowlist{
		UvBinary:         "uv",
		WorkspaceRootAbs: ws,
		PolicyFileAbs:    policyPath,
		PolicyID:         id,
		Symbol:           sym,
	}

	argv := []string{
		"uv", "run", "python", "-m", "iterative_context.validate_policy",
		"--policy-path", policyPath,
		"--policy-id", id,
		"--symbol", sym,
		"--json",
	}
	if !al.Allows(argv) {
		t.Fatal("expected validate_policy argv to be allowlisted")
	}
}

func TestICOptimizerAllowlistRejectsPolicyPathOutsideWorkspace(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	outside := filepath.Join(t.TempDir(), "evil.py")
	al := ICOptimizerAllowlist{
		UvBinary:         "uv",
		WorkspaceRootAbs: tmp,
		PolicyFileAbs:    outside,
		PolicyID:         "id",
		Symbol:           "score_fn",
	}
	argv := []string{
		"uv", "run", "python", "-m", "iterative_context.validate_policy",
		"--policy-path", outside,
		"--policy-id", "id",
		"--symbol", "score_fn",
		"--json",
	}
	if al.Allows(argv) {
		t.Fatal("expected argv outside workspace root to be rejected")
	}
}

func TestICOptimizerAllowlistRejectsArbitraryShell(t *testing.T) {
	t.Parallel()

	al := ICOptimizerAllowlist{UvBinary: "uv", WorkspaceRootAbs: "/tmp", PolicyFileAbs: "/tmp/p.py", PolicyID: "x", Symbol: "score_fn"}
	if al.Allows([]string{"sh", "-c", "pytest"}) {
		t.Fatal("expected sh -c to be rejected")
	}
}
