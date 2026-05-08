package config

import (
	"bytes"
	"context"
	"os"
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

	if experiment.Name != "local-ic-vs-jcodemunch-round-001" {
		t.Fatalf("experiment.Name = %q", experiment.Name)
	}
	if experiment.Mode != ModeEvaluation {
		t.Fatalf("experiment.Mode = %q, want %q", experiment.Mode, ModeEvaluation)
	}
	if experiment.Agents.Evaluator == nil {
		t.Fatal("experiment.Agents.Evaluator is nil")
	}
	if experiment.Agents.Evaluator.Model.Provider != ProviderFake {
		t.Fatalf("experiment.Agents.Evaluator.Model.Provider = %q, want %q", experiment.Agents.Evaluator.Model.Provider, ProviderFake)
	}
	if experiment.Evaluation == nil {
		t.Fatal("experiment.Evaluation is nil")
	}
	if experiment.Evaluation.Scoring.Objective != "scoring/localization-objective.pkl" {
		t.Fatalf("experiment.Evaluation.Scoring.Objective = %q", experiment.Evaluation.Scoring.Objective)
	}
	if experiment.Artifacts.CandidatePolicyRound001 == nil || experiment.Artifacts.CandidatePolicyRound001.Path != "policies/candidate_policy.py" {
		t.Fatalf("candidate policy artifact = %#v, want local policy path", experiment.Artifacts.CandidatePolicyRound001)
	}
}

func TestLoadOptimizeICManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	path := filepath.Join("..", "..", "..", "..", "configs", "experiments", "optimize-ic", "experiment.pkl")
	experiment, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if experiment.Name != "optimize-ic-round-002" {
		t.Fatalf("experiment.Name = %q", experiment.Name)
	}
	if experiment.Mode != ModeOptimization {
		t.Fatalf("experiment.Mode = %q, want %q", experiment.Mode, ModeOptimization)
	}
	if experiment.Agents.Optimizer == nil {
		t.Fatal("experiment.Agents.Optimizer is nil")
	}
	if experiment.Optimization == nil {
		t.Fatal("experiment.Optimization is nil")
	}
	if experiment.Evaluation != nil {
		t.Fatalf("experiment.Evaluation = %#v, want nil for optimization manifest", experiment.Evaluation)
	}
	if experiment.Artifacts.CandidatePolicyRound001 == nil || experiment.Artifacts.CandidatePolicyRound001.Path != "policies/candidate_policy.py" {
		t.Fatalf("experiment.Artifacts.CandidatePolicyRound001 = %#v, want local optimize policy path", experiment.Artifacts.CandidatePolicyRound001)
	}
	if experiment.Optimization.Target.Input.Id != "candidate-policy-round-001" {
		t.Fatalf("experiment.Optimization.Target.Input.Id = %q", experiment.Optimization.Target.Input.Id)
	}
	if experiment.Optimization.Target.Output.ArtifactName != "candidate_policy.round-002.py" {
		t.Fatalf("experiment.Optimization.Target.Output.ArtifactName = %q", experiment.Optimization.Target.Output.ArtifactName)
	}
}

func TestOptimizeICExampleSupportFilesMatchEvaluationExample(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..", "..", "..")
	localPolicyPath := filepath.Join(repoRoot, "configs", "experiments", "local-ic-vs-jcodemunch", "policies", "candidate_policy.py")
	optimizePolicyPath := filepath.Join(repoRoot, "configs", "experiments", "optimize-ic", "policies", "candidate_policy.py")
	localObjectivePath := filepath.Join(repoRoot, "configs", "experiments", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl")
	optimizeObjectivePath := filepath.Join(repoRoot, "configs", "experiments", "optimize-ic", "scoring", "localization-objective.pkl")

	localPolicy, err := os.ReadFile(localPolicyPath)
	if err != nil {
		t.Fatalf("os.ReadFile(local policy) error = %v", err)
	}
	optimizePolicy, err := os.ReadFile(optimizePolicyPath)
	if err != nil {
		t.Fatalf("os.ReadFile(optimize policy) error = %v", err)
	}
	if !bytes.Equal(localPolicy, optimizePolicy) {
		t.Fatalf("optimize policy fixture drifted from local evaluation policy")
	}

	localObjective, err := os.ReadFile(localObjectivePath)
	if err != nil {
		t.Fatalf("os.ReadFile(local objective) error = %v", err)
	}
	optimizeObjective, err := os.ReadFile(optimizeObjectivePath)
	if err != nil {
		t.Fatalf("os.ReadFile(optimize objective) error = %v", err)
	}
	if !bytes.Equal(localObjective, optimizeObjective) {
		t.Fatalf("optimize scoring fixture drifted from local evaluation scoring")
	}
}

func requirePkl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pkl"); err != nil {
		t.Skip("pkl CLI not available on PATH")
	}
}
