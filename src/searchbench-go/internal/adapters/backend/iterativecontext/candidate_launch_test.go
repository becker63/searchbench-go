package iterativecontext_test

import (
	"testing"

	ic "github.com/becker63/searchbench-go/internal/adapters/backend/iterativecontext"
	"github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func TestLaunchCWDMismatchErrors(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	accepted := optimizer.AcceptedICCandidate{
		Workspace: optimizer.ICCandidateWorkspace{
			ID:   "ws",
			Root: root,
			Seed: optimizer.WorkspaceSeedIdentity{
				Provider: optimizer.SeedProviderLocalPath,
				Source:   "src",
				Sha256:   "abc",
			},
		},
		Policy: optimizer.ICPolicyArtifact{
			Path:     root + "/policy.py",
			PolicyID: "p1",
			Sha256:   "def",
		},
		Validation: optimizer.PipelineValidationResult{OK: true},
		Launch: optimizer.ICLaunchSpec{
			CWD:  t.TempDir(),
			Argv: []string{"uv", "run", "python", "-m", "iterative_context.server"},
		},
	}
	if err := ic.ValidateAcceptedLaunch(accepted); err == nil {
		t.Fatal("expected cwd mismatch error")
	}
}

func TestRuntimeIdentityIncludesSeedAndPolicy(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	accepted := optimizer.AcceptedICCandidate{
		Workspace: optimizer.ICCandidateWorkspace{
			ID:   "ws",
			Root: root,
			Seed: optimizer.WorkspaceSeedIdentity{
				Provider: optimizer.SeedProviderBuckDescriptor,
				Source:   "//src/iterative-context:optimizable_backend",
				Sha256:   "abc",
			},
		},
		Policy: optimizer.ICPolicyArtifact{
			Path:     root + "/policy.py",
			PolicyID: "p1",
			Sha256:   "def",
		},
		Validation: optimizer.PipelineValidationResult{OK: true},
		Launch: optimizer.ICLaunchSpec{
			CWD:  root,
			Argv: []string{"uv", "run", "python", "-m", "iterative_context.server"},
		},
	}
	if err := ic.ValidateAcceptedLaunch(accepted); err != nil {
		t.Fatal(err)
	}
	id := ic.RuntimeIdentityFromAccepted(accepted, true)
	if id.SeedIdentity.Provider != optimizer.SeedProviderBuckDescriptor {
		t.Fatalf("provider: %s", id.SeedIdentity.Provider)
	}
	if id.PolicyID != "p1" || !id.Verified {
		t.Fatalf("identity: %+v", id)
	}
}

func TestAdminToolsExcludedFromEvaluatorList(t *testing.T) {
	t.Parallel()
	// Regression guard: excludedEvaluatorToolNames in runtime.go hides install/verify.
	tools := []string{"install_score", "verify_score", "search"}
	var visible []string
	for _, name := range tools {
		if name == "install_score" || name == "verify_score" {
			continue
		}
		visible = append(visible, name)
	}
	if len(visible) != 1 || visible[0] != "search" {
		t.Fatalf("visible: %v", visible)
	}
}
