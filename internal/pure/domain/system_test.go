package domain

import "testing"

func TestPolicyArtifactValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		policy  PolicyArtifact
		wantErr bool
	}{
		{
			name:   "valid python policy",
			policy: NewPythonPolicy(PolicyID("policy-1"), "def run():\n    return 1\n", "run"),
		},
		{
			name: "mismatched hash",
			policy: PolicyArtifact{
				ID:         PolicyID("policy-1"),
				Language:   PolicyLanguagePython,
				SHA256:     PolicyHash("deadbeef"),
				Entrypoint: "run",
				Source:     "def run():\n    return 1\n",
			},
			wantErr: true,
		},
		{
			name: "missing entrypoint",
			policy: PolicyArtifact{
				ID:       PolicyID("policy-1"),
				Language: PolicyLanguagePython,
				SHA256:   NewPythonPolicy(PolicyID("policy-1"), "print('x')", "run").SHA256,
				Source:   "print('x')",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.policy.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSystemSpecFingerprintDeterminism(t *testing.T) {
	t.Parallel()

	policy := NewPythonPolicy(PolicyID("policy-1"), "def choose():\n    return 'ok'\n", "choose")
	base := SystemSpec{
		ID:      SystemID("system-1"),
		Name:    "Baseline Name",
		Backend: BackendIterativeContext,
		Model: ModelSpec{
			Provider: "openai",
			Name:     "gpt-test",
		},
		PromptBundle: PromptBundleRef{
			Name:    "default",
			Version: "v1",
		},
		Policy: &policy,
		Runtime: RuntimeConfig{
			MaxSteps:        12,
			MaxOutputTokens: 500,
		},
	}

	same := base
	if got, want := base.Fingerprint(), same.Fingerprint(); got != want {
		t.Fatalf("fingerprint mismatch for identical system: got %q want %q", got, want)
	}

	renamed := base
	renamed.Name = "Candidate Display Name"
	if got, want := base.Fingerprint(), renamed.Fingerprint(); got != want {
		t.Fatalf("fingerprint should ignore cosmetic name: got %q want %q", got, want)
	}

	changedRuntime := base
	changedRuntime.Runtime.MaxSteps = 99
	if base.Fingerprint() == changedRuntime.Fingerprint() {
		t.Fatal("fingerprint should change when behavior-affecting fields change")
	}

	changedEntrypoint := base
	changedPolicy := *base.Policy
	changedPolicy.Entrypoint = "alternate"
	changedEntrypoint.Policy = &changedPolicy
	if base.Fingerprint() == changedEntrypoint.Fingerprint() {
		t.Fatal("fingerprint should change when policy entrypoint changes")
	}
}

func TestSystemSpecRefOmitsPolicySource(t *testing.T) {
	t.Parallel()

	policy := NewPythonPolicy(PolicyID("policy-1"), "def choose():\n    return 'ok'\n", "choose")
	spec := SystemSpec{
		ID:      SystemID("system-1"),
		Name:    "Candidate",
		Backend: BackendIterativeContext,
		Model: ModelSpec{
			Provider: "openai",
			Name:     "gpt-test",
		},
		PromptBundle: PromptBundleRef{Name: "default"},
		Policy:       &policy,
	}

	ref := spec.Ref()
	if ref.Policy == nil {
		t.Fatal("expected policy ref")
	}
	if ref.Policy.Entrypoint != policy.Entrypoint {
		t.Fatalf("Policy.Entrypoint = %q, want %q", ref.Policy.Entrypoint, policy.Entrypoint)
	}
	if ref.Fingerprint != spec.Fingerprint() {
		t.Fatalf("Fingerprint = %q, want %q", ref.Fingerprint, spec.Fingerprint())
	}
}
