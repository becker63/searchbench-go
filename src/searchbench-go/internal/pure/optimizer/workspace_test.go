package optimizer

import (
	"testing"
)

func TestWorkspaceSeedIdentityValidate(t *testing.T) {
	t.Parallel()
	valid := WorkspaceSeedIdentity{
		Provider: SeedProviderLocalPath,
		Source:   "/tmp/src",
		Sha256:   "abc",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid identity: %v", err)
	}
	if err := (WorkspaceSeedIdentity{}).Validate(); err == nil {
		t.Fatal("expected error for empty identity")
	}
}

func TestWorkspaceSeedValidate(t *testing.T) {
	t.Parallel()
	seed := WorkspaceSeed{
		ID:   "seed-1",
		Kind: SeedKindLocalPath,
		Root: "/data/ic",
		Identity: WorkspaceSeedIdentity{
			Provider: SeedProviderLocalPath,
			Source:   "src/iterative-context",
			Sha256:   "deadbeef",
		},
	}
	if err := seed.Validate(); err != nil {
		t.Fatalf("seed validate: %v", err)
	}
	buck := seed
	buck.Kind = SeedKindBuckDescriptor
	buck.Identity.Provider = SeedProviderBuckDescriptor
	buck.Identity.Source = "//src/iterative-context:optimizable_backend"
	if err := buck.Validate(); err != nil {
		t.Fatalf("buck seed validate: %v", err)
	}
}

func TestAcceptedICCandidateLaunchCWDMustMatchWorkspace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	candidate := AcceptedICCandidate{
		Workspace: ICCandidateWorkspace{
			ID:   "ws-1",
			Root: root,
			Seed: WorkspaceSeedIdentity{
				Provider: SeedProviderLocalPath,
				Source:   "src/iterative-context",
				Sha256:   "abc",
			},
		},
		Policy: ICPolicyArtifact{
			Path:     root + "/policy.py",
			PolicyID: "pol-1",
			Symbol:   "score_fn",
			Sha256:   "def",
		},
		Validation: PipelineValidationResult{OK: true},
		Launch: ICLaunchSpec{
			CWD:  "/other",
			Argv: []string{"uv", "run"},
		},
	}
	if err := candidate.Validate(); err == nil {
		t.Fatal("expected cwd mismatch error")
	}
	candidate.Launch.CWD = root
	if err := candidate.Validate(); err != nil {
		t.Fatalf("expected valid accepted candidate: %v", err)
	}
}

func TestICPolicyArtifactFromStagedMeta(t *testing.T) {
	t.Parallel()
	artifact := ICPolicyArtifactFromStagedMeta(
		"/ws/policy.py", "artifact-1", "score_fn", "iface-1", "def score_fn():\n  pass\n",
	)
	if artifact.PolicyID != "artifact-1" || artifact.Sha256 == "" {
		t.Fatalf("unexpected artifact: %+v", artifact)
	}
}
