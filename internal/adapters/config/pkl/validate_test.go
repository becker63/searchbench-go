package config

import (
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/testing/importcheck"
)

func TestValidFromScratchRoundValidates(t *testing.T) {
	t.Parallel()

	if err := Validate(sampleFromScratchRound()); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidContinuationRoundValidates(t *testing.T) {
	t.Parallel()

	if err := Validate(sampleContinuationRound()); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestMissingRoundBlockFails(t *testing.T) {
	t.Parallel()

	spec := sampleFromScratchRound()
	spec.Round = nil

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrMissingRoundManifest.Error()) {
		t.Fatalf("Validate() error = %v, want missing round block error", err)
	}
}

func TestFromScratchRequiresIncumbent(t *testing.T) {
	t.Parallel()

	spec := sampleFromScratchRound()
	spec.Round.Incumbent = nil

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrMissingRoundIncumbent.Error()) {
		t.Fatalf("Validate() error = %v, want missing incumbent error", err)
	}
}

func TestFromScratchRequiresMatches(t *testing.T) {
	t.Parallel()

	spec := sampleFromScratchRound()
	spec.Round.Matches = nil

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrMissingRoundMatches.Error()) {
		t.Fatalf("Validate() error = %v, want missing matches error", err)
	}
}

func TestFromScratchRequiresEvaluator(t *testing.T) {
	t.Parallel()

	spec := sampleFromScratchRound()
	spec.Round.Evaluator = nil

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrMissingRoundEvaluator.Error()) {
		t.Fatalf("Validate() error = %v, want missing evaluator error", err)
	}
}

func TestFromScratchRequiresScoring(t *testing.T) {
	t.Parallel()

	spec := sampleFromScratchRound()
	spec.Round.Scoring = nil

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrMissingRoundScoring.Error()) {
		t.Fatalf("Validate() error = %v, want missing scoring error", err)
	}
}

func TestContinuationRejectsMissingChallengerPatch(t *testing.T) {
	t.Parallel()

	spec := sampleContinuationRound()
	spec.Round.Challenger.SelectionPolicy = nil

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrMissingRoundChallenger.Error()) {
		t.Fatalf("Validate() error = %v, want missing challenger patch error", err)
	}
}

func TestContinuationRejectsConflictingChallengerPatch(t *testing.T) {
	t.Parallel()

	spec := sampleContinuationRound()
	spec.Round.Challenger.Generate = &GeneratedChallenger{
		Optimizer: sampleOptimizer(),
	}

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrRoundChallengerConflict.Error()) {
		t.Fatalf("Validate() error = %v, want challenger conflict error", err)
	}
}

func TestGeneratedArtifactNameMustBeRelative(t *testing.T) {
	t.Parallel()

	spec := sampleGeneratedContinuationRound()
	spec.Round.Challenger.Generate.ArtifactName = "../bad.py"

	if err := Validate(spec); err == nil || !strings.Contains(err.Error(), ErrRoundGenerateArtifactNameInvalid.Error()) {
		t.Fatalf("Validate() error = %v, want invalid generated artifact name error", err)
	}
}

func TestPurePackagesDoNotImportPkl(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".."))
	dirs := []string{
		filepath.Join(repoRoot, "internal", "pure", "domain"),
		filepath.Join(repoRoot, "internal", "pure", "execution"),
		filepath.Join(repoRoot, "internal", "pure", "score"),
		filepath.Join(repoRoot, "internal", "pure", "report"),
		filepath.Join(repoRoot, "internal", "pure", "codegraph"),
		filepath.Join(repoRoot, "internal", "agents", "evaluator", "prompt"),
		filepath.Join(repoRoot, "internal", "agents", "optimizer", "prompt"),
		filepath.Join(repoRoot, "internal", "pure", "usage"),
	}

	for _, dir := range dirs {
		fs := token.NewFileSet()
		files, err := importcheck.ParseNonTestGoFilesImportsOnly(fs, dir)
		if err != nil {
			t.Fatalf("ParseNonTestGoFilesImportsOnly(%q) error = %v", dir, err)
		}
		for _, file := range files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.Contains(path, "github.com/apple/pkl-go") {
					t.Fatalf("pure package import %q leaked pkl-go", path)
				}
			}
		}
	}
}

func sampleFromScratchRound() RoundSpec {
	prompt := "Use structural code evidence before guessing."
	spec := RoundSpec{
		Name: "local-ic-vs-jcodemunch-round-001",
		Interfaces: Interfaces{
			IterativeContextSelectionPolicyV1: Interface{Id: "iterative_context.selection_policy.v1"},
		},
		Round: &RoundManifest{
			Id: "round-001",
			Incumbent: &RoundPolicy{
				System: System{
					Id:      "jcodemunch",
					Name:    "jCodeMunch",
					Backend: BackendJCodeMunch,
				},
			},
			Challenger: RoundChallenger{
				System: &System{
					Id:      "iterative-context",
					Name:    "Iterative Context",
					Backend: BackendIterativeContext,
				},
				SelectionPolicy: &PolicyArtifact{
					Id:   "challenger-policy-round-001",
					Kind: "policy",
					Path: "policies/challenger_policy.py",
					Implements: Interface{
						Id: "iterative_context.selection_policy.v1",
					},
				},
			},
			Matches: &Dataset{
				Kind:     "lca",
				Name:     "JetBrains-Research/lca-bug-localization",
				Config:   "py",
				Split:    "dev",
				MaxItems: intPtr(5),
			},
			Evaluator: &Evaluator{
				Model: sampleModel("fake-evaluator"),
				Bounds: AgentBounds{
					MaxModelTurns:  8,
					MaxToolCalls:   24,
					TimeoutSeconds: 300,
				},
				Tools: AgentToolPolicy{
					Allow: []string{"resolve", "expand", "resolve_and_expand"},
					Deny:  []string{"shell", "go_test", "write_file", "network"},
				},
				SystemPrompt: &prompt,
				Retry: RetryPolicy{
					MaxAttempts:                2,
					RetryOnModelError:          true,
					RetryOnToolFailure:         false,
					RetryOnFinalizationFailure: true,
					RetryOnInvalidPrediction:   true,
				},
			},
			Scoring: &Scoring{
				Objective: "scoring/localization-objective.pkl",
			},
			Report: Report{
				Formats: []ReportFormat{ReportFormatJSON, ReportFormatText},
			},
		},
	}
	return spec
}

func sampleContinuationRound() RoundSpec {
	spec := sampleFromScratchRound()
	spec.Name = "continue-ic-from-local-round-002"
	spec.Round = &RoundManifest{
		Id:        "round-002",
		Continues: stringPtr("../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001"),
		Challenger: RoundChallenger{
			SelectionPolicy: &PolicyArtifact{
				Id:   "challenger-policy-round-002",
				Kind: "policy",
				Path: "policies/challenger_policy.py",
				Implements: Interface{
					Id: "iterative_context.selection_policy.v1",
				},
			},
		},
	}
	return spec
}

func sampleGeneratedContinuationRound() RoundSpec {
	spec := sampleContinuationRound()
	spec.Name = "generate-ic-from-local-round-002"
	spec.Round.Challenger.SelectionPolicy = nil
	spec.Round.Challenger.Generate = &GeneratedChallenger{
		Optimizer: sampleOptimizer(),
	}
	return spec
}

func sampleOptimizer() Optimizer {
	prompt := "Improve the Iterative Context selection policy."
	return Optimizer{
		Model: Model{
			Provider:        ProviderFake,
			Name:            "fake-optimizer",
			MaxOutputTokens: intPtr(4000),
		},
		Bounds: AgentBounds{
			MaxModelTurns:  8,
			MaxToolCalls:   16,
			TimeoutSeconds: 300,
		},
		Tools: AgentToolPolicy{
			Allow: []string{"read_parent_bundle", "read_objective_result"},
			Deny:  []string{"shell", "network"},
		},
		SystemPrompt: &prompt,
	}
}

func sampleModel(name string) Model {
	return Model{
		Provider:        ProviderFake,
		Name:            name,
		MaxOutputTokens: intPtr(2000),
	}
}

func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
