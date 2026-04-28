package config

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoadLocalICVsJCodeMunchManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	path := filepath.Join("..", "..", "..", "..", "configs", "experiments", "local-ic-vs-jcodemunch", "experiment.pkl")
	experiment, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if experiment.Name != "local-ic-vs-jcodemunch-lca-dev" {
		t.Fatalf("experiment.Name = %q", experiment.Name)
	}
	if experiment.Mode != ModeEvaluatorOnly {
		t.Fatalf("experiment.Mode = %q, want %q", experiment.Mode, ModeEvaluatorOnly)
	}
	if experiment.Writer != nil {
		t.Fatalf("experiment.Writer = %#v, want nil for evaluator_only example", experiment.Writer)
	}
	if experiment.Evaluator.Model.Provider != ProviderOpenRouter {
		t.Fatalf("experiment.Evaluator.Model.Provider = %q, want %q", experiment.Evaluator.Model.Provider, ProviderOpenRouter)
	}
	if experiment.Scoring.Objective != "scoring/localization-objective.pkl" {
		t.Fatalf("experiment.Scoring.Objective = %q", experiment.Scoring.Objective)
	}
	if experiment.Systems.Candidate.Policy == nil || experiment.Systems.Candidate.Policy.Path != "policies/candidate_policy.py" {
		t.Fatalf("candidate policy path = %#v, want local policy path", experiment.Systems.Candidate.Policy)
	}
	if experiment.OutputConfig.BundleRoot != "artifacts/runs" {
		t.Fatalf("experiment.OutputConfig.BundleRoot = %q, want local bundle root", experiment.OutputConfig.BundleRoot)
	}
}

func requirePkl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
}
