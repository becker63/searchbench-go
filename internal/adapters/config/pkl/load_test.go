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

	path := filepath.Join("..", "..", "..", "..", "configs", "rounds", "local-ic-vs-jcodemunch", "round.pkl")
	roundSpec, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if roundSpec.Name != "local-ic-vs-jcodemunch-round-001" {
		t.Fatalf("roundSpec.Name = %q", roundSpec.Name)
	}
	if roundSpec.Mode != ModeEvaluation {
		t.Fatalf("roundSpec.Mode = %q, want %q", roundSpec.Mode, ModeEvaluation)
	}
	if roundSpec.Agents.Evaluator == nil {
		t.Fatal("roundSpec.Agents.Evaluator is nil")
	}
	if roundSpec.Agents.Evaluator.Model.Provider != ProviderFake {
		t.Fatalf("roundSpec.Agents.Evaluator.Model.Provider = %q, want %q", roundSpec.Agents.Evaluator.Model.Provider, ProviderFake)
	}
	if roundSpec.Evaluation == nil {
		t.Fatal("roundSpec.Evaluation is nil")
	}
	if roundSpec.Evaluation.Scoring.Objective != "scoring/localization-objective.pkl" {
		t.Fatalf("roundSpec.Evaluation.Scoring.Objective = %q", roundSpec.Evaluation.Scoring.Objective)
	}
	if roundSpec.Artifacts.ChallengerPolicyRound001 == nil || roundSpec.Artifacts.ChallengerPolicyRound001.Path != "policies/challenger_policy.py" {
		t.Fatalf("challenger policy artifact = %#v, want local policy path", roundSpec.Artifacts.ChallengerPolicyRound001)
	}
}

func TestLoadOptimizeICManifest(t *testing.T) {
	t.Parallel()

	requirePkl(t)

	path := filepath.Join("..", "..", "..", "..", "configs", "rounds", "optimize-ic", "round.pkl")
	roundSpec, err := LoadFromPath(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadFromPath() error = %v", err)
	}

	if roundSpec.Name != "optimize-ic-round-002" {
		t.Fatalf("roundSpec.Name = %q", roundSpec.Name)
	}
	if roundSpec.Mode != ModeOptimization {
		t.Fatalf("roundSpec.Mode = %q, want %q", roundSpec.Mode, ModeOptimization)
	}
	if roundSpec.Agents.Optimizer == nil {
		t.Fatal("roundSpec.Agents.Optimizer is nil")
	}
	if roundSpec.Optimization == nil {
		t.Fatal("roundSpec.Optimization is nil")
	}
	if roundSpec.Evaluation != nil {
		t.Fatalf("roundSpec.Evaluation = %#v, want nil for optimization manifest", roundSpec.Evaluation)
	}
	if roundSpec.Artifacts.ChallengerPolicyRound001 == nil || roundSpec.Artifacts.ChallengerPolicyRound001.Path != "policies/challenger_policy.py" {
		t.Fatalf("roundSpec.Artifacts.ChallengerPolicyRound001 = %#v, want local optimize policy path", roundSpec.Artifacts.ChallengerPolicyRound001)
	}
	if roundSpec.Optimization.Target.Input.Id != "challenger-policy-round-001" {
		t.Fatalf("roundSpec.Optimization.Target.Input.Id = %q", roundSpec.Optimization.Target.Input.Id)
	}
	if roundSpec.Optimization.Target.Output.ArtifactName != "next_challenger_policy.round-002.py" {
		t.Fatalf("roundSpec.Optimization.Target.Output.ArtifactName = %q", roundSpec.Optimization.Target.Output.ArtifactName)
	}
}

func TestOptimizeICExampleSupportFilesMatchEvaluationExample(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..", "..", "..")
	localPolicyPath := filepath.Join(repoRoot, "configs", "rounds", "local-ic-vs-jcodemunch", "policies", "challenger_policy.py")
	optimizePolicyPath := filepath.Join(repoRoot, "configs", "rounds", "optimize-ic", "policies", "challenger_policy.py")
	localObjectivePath := filepath.Join(repoRoot, "configs", "rounds", "local-ic-vs-jcodemunch", "scoring", "localization-objective.pkl")
	optimizeObjectivePath := filepath.Join(repoRoot, "configs", "rounds", "optimize-ic", "scoring", "localization-objective.pkl")

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
