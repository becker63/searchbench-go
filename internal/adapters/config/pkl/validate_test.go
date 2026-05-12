package config

import (
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/becker63/searchbench-go/internal/testing/importcheck"
)

func TestValidEvaluationConfigValidates(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	if err := Validate(roundSpec); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidOptimizationConfigValidates(t *testing.T) {
	t.Parallel()

	roundSpec := sampleOptimizationRound()
	if roundSpec.Evaluation != nil {
		t.Fatalf("sampleOptimizationRound().Evaluation = %#v, want nil", roundSpec.Evaluation)
	}
	if err := Validate(roundSpec); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestOptimizationModeDoesNotRequireEvaluationOrScoring(t *testing.T) {
	t.Parallel()

	roundSpec := sampleOptimizationRound()
	roundSpec.Evaluation = nil

	if err := Validate(roundSpec); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestEvaluationModeRequiresEvaluation(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Evaluation = nil

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingEvaluation.Error()) {
		t.Fatalf("Validate() error = %v, want missing evaluation error", err)
	}
}

func TestEvaluationModeRequiresEvaluator(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Agents.Evaluator = nil

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingEvaluator.Error()) {
		t.Fatalf("Validate() error = %v, want missing evaluator error", err)
	}
}

func TestOptimizationModeRequiresOptimization(t *testing.T) {
	t.Parallel()

	roundSpec := sampleOptimizationRound()
	roundSpec.Optimization = nil

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingOptimization.Error()) {
		t.Fatalf("Validate() error = %v, want missing optimization error", err)
	}
}

func TestOptimizationModeRequiresOptimizer(t *testing.T) {
	t.Parallel()

	roundSpec := sampleOptimizationRound()
	roundSpec.Agents.Optimizer = nil

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingOptimizer.Error()) {
		t.Fatalf("Validate() error = %v, want missing optimizer error", err)
	}
}

func TestRoundConfigNoLongerHasWriterOrOutputConfig(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(RoundSpec{})
	if _, ok := typ.FieldByName("Writer"); ok {
		t.Fatal("RoundSpec unexpectedly has Writer field")
	}
	if _, ok := typ.FieldByName("OutputConfig"); ok {
		t.Fatal("RoundSpec unexpectedly has OutputConfig field")
	}
}

func TestRejectsMissingDatasetConfig(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Dataset.Config = ""

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingDatasetConfig.Error()) {
		t.Fatalf("Validate() error = %v, want missing dataset config error", err)
	}
}

func TestRejectsMissingDatasetSplit(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Dataset.Split = ""

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingDatasetSplit.Error()) {
		t.Fatalf("Validate() error = %v, want missing dataset split error", err)
	}
}

func TestRejectsMissingIncumbentPolicyID(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Policies.Incumbent.Id = ""

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingIncumbentPolicyID.Error()) {
		t.Fatalf("Validate() error = %v, want missing incumbent id error", err)
	}
}

func TestRejectsMissingIterativeContextSystemID(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Policies.Challenger.Id = ""

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingChallengerPolicyID.Error()) {
		t.Fatalf("Validate() error = %v, want missing iterative context id error", err)
	}
}

func TestRejectsMissingAgentModelProvider(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Agents.Evaluator.Model.Provider = ""

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingAgentModelProvider.Error()) {
		t.Fatalf("Validate() error = %v, want missing provider error", err)
	}
}

func TestRejectsMissingAgentModelName(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Agents.Evaluator.Model.Name = ""

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingAgentModelName.Error()) {
		t.Fatalf("Validate() error = %v, want missing model name error", err)
	}
}

func TestRejectsMissingScoringObjectivePath(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Evaluation.Scoring.Objective = ""

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrMissingScoringObjectivePath.Error()) {
		t.Fatalf("Validate() error = %v, want missing scoring objective error", err)
	}
}

func TestRejectsEvaluationAgentMismatch(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Evaluation.Agent.Model.Name = "other"

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrEvaluationAgentMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want evaluation agent mismatch", err)
	}
}

func TestRejectsEvaluationChallengerPolicyMismatch(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Evaluation.Challenger.System.Id = "other"

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrEvaluationChallengerPolicyMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want challenger policy mismatch", err)
	}
}

func TestRejectsSelectionPolicyArtifactMismatch(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Evaluation.Challenger.Uses.SelectionPolicy.Id = "other"

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrChallengerSelectionPolicyArtifactMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want selection policy mismatch", err)
	}
}

func TestRejectsSelectionPolicyInterfaceMismatch(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Artifacts.ChallengerPolicy.Implements.Id = "other"
	roundSpec.Evaluation.Challenger.Uses.SelectionPolicy.Implements.Id = "other"

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrChallengerSelectionPolicyInterfaceMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want selection policy interface mismatch", err)
	}
}

func TestRejectsAbsolutePolicyArtifactPath(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Artifacts.ChallengerPolicy.Path = "/tmp/policy.py"

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrPolicyArtifactPathMustBeRelative.Error()) {
		t.Fatalf("Validate() error = %v, want relative policy path error", err)
	}
}

func TestRejectsProposalArtifactNameWithParentTraversal(t *testing.T) {
	t.Parallel()

	roundSpec := sampleOptimizationRound()
	roundSpec.Artifacts.NextChallenger.ArtifactName = "../challenger.py"

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrNextChallengerArtifactNameInvalid.Error()) {
		t.Fatalf("Validate() error = %v, want invalid artifact name error", err)
	}
}

func TestRejectsNextChallengerTargetOutputMismatch(t *testing.T) {
	t.Parallel()

	roundSpec := sampleOptimizationRound()
	roundSpec.Optimization.Target.Output.Id = "other"

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrOptimizationNextChallengerOutputMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want optimization target output mismatch", err)
	}
}

func TestRejectsEmptyToolAllowEntry(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Agents.Evaluator.Tools.Allow = append(roundSpec.Agents.Evaluator.Tools.Allow, " ")

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrToolAllowEntryEmpty.Error()) {
		t.Fatalf("Validate() error = %v, want empty allow entry error", err)
	}
}

func TestRejectsDuplicateToolAllowEntry(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Agents.Evaluator.Tools.Allow = append(roundSpec.Agents.Evaluator.Tools.Allow, roundSpec.Agents.Evaluator.Tools.Allow[0])

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrToolAllowDuplicate.Error()) {
		t.Fatalf("Validate() error = %v, want duplicate allow entry error", err)
	}
}

func TestRejectsToolAllowDenyOverlap(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Agents.Evaluator.Tools.Deny = append(roundSpec.Agents.Evaluator.Tools.Deny, roundSpec.Agents.Evaluator.Tools.Allow[0])

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrToolPolicyOverlap.Error()) {
		t.Fatalf("Validate() error = %v, want tool overlap error", err)
	}
}

func TestRejectsOversizedSystemPrompt(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	oversized := strings.Repeat("x", maxSystemPromptBytes+1)
	roundSpec.Agents.Evaluator.SystemPrompt = &oversized
	roundSpec.Evaluation.Agent.SystemPrompt = &oversized

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrSystemPromptTooLarge.Error()) {
		t.Fatalf("Validate() error = %v, want oversized prompt error", err)
	}
}

func TestRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	roundSpec := sampleEvaluationRound()
	roundSpec.Mode = RunMode("mystery_mode")

	if err := Validate(roundSpec); err == nil || !strings.Contains(err.Error(), ErrUnsupportedMode.Error()) {
		t.Fatalf("Validate() error = %v, want unsupported mode error", err)
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

func sampleEvaluationRound() RoundSpec {
	prompt := "Use structural code evidence before guessing."
	evaluator := &Evaluator{
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
	}

	roundSpec := RoundSpec{
		Name: "local-ic-vs-jcodemunch-round-001",
		Mode: ModeEvaluation,
		Dataset: Dataset{
			Kind:     "lca",
			Name:     "JetBrains-Research/lca-bug-localization",
			Config:   "py",
			Split:    "dev",
			MaxItems: intPtr(5),
		},
		Interfaces: Interfaces{
			IterativeContextSelectionPolicyV1: Interface{Id: "iterative_context.selection_policy.v1"},
		},
		Policies: Policies{
			Incumbent: System{
				Id:      "jcodemunch",
				Name:    "jCodeMunch",
				Backend: BackendJCodeMunch,
				PromptBundle: PromptBundle{
					Name:    "default",
					Version: stringPtr("dev"),
				},
				Runtime: Runtime{
					MaxSteps:       16,
					TimeoutSeconds: 300,
				},
			},
			Challenger: System{
				Id:      "iterative-context",
				Name:    "Iterative Context",
				Backend: BackendIterativeContext,
				PromptBundle: PromptBundle{
					Name:    "default",
					Version: stringPtr("dev"),
				},
				Runtime: Runtime{
					MaxSteps:       16,
					TimeoutSeconds: 300,
				},
			},
		},
		Artifacts: Artifacts{
			ChallengerPolicy: &PolicyArtifact{
				Id:   "challenger-policy-round-001",
				Kind: "policy",
				Path: "policies/challenger_policy.py",
				Implements: Interface{
					Id: "iterative_context.selection_policy.v1",
				},
			},
		},
		Agents: Agents{
			Evaluator: evaluator,
		},
	}

	roundSpec.Evaluation = &Evaluation{
		Agent: *evaluator,
		Incumbent: EvaluationSystemBinding{
			System: roundSpec.Policies.Incumbent,
		},
		Challenger: ChallengerEvaluationBinding{
			System: roundSpec.Policies.Challenger,
			Uses: ChallengerUses{
				SelectionPolicy: *roundSpec.Artifacts.ChallengerPolicy,
			},
		},
		Scoring: Scoring{
			Objective: "scoring/localization-objective.pkl",
		},
		Report: Report{
			Formats: []ReportFormat{ReportFormatJSON, ReportFormatText},
		},
	}

	return roundSpec
}

func sampleOptimizationRound() RoundSpec {
	roundSpec := sampleEvaluationRound()
	prompt := "Improve the Iterative Context selection policy using only the provided parent-run evidence."
	optimizer := &Optimizer{
		Model: sampleModel("fake-optimizer"),
		Bounds: AgentBounds{
			MaxModelTurns:  8,
			MaxToolCalls:   16,
			TimeoutSeconds: 300,
		},
		Tools: AgentToolPolicy{
			Allow: []string{
				"read_parent_bundle",
				"read_round_evidence",
				"read_objective_result",
				"read_report_summary",
				"read_artifact",
				"write_artifact_proposal",
			},
			Deny: []string{"run_evaluator", "shell", "go_test", "write_manifest", "network"},
		},
		SystemPrompt: &prompt,
	}

	roundSpec.Name = "optimize-ic-round-002"
	roundSpec.Mode = ModeOptimization
	roundSpec.Evaluation = nil
	roundSpec.Artifacts.ParentRoundBundle = &CompletedRoundBundleArtifact{
		Id:   "local-ic-vs-jcodemunch-round-001",
		Kind: "completed_round_bundle",
		Path: "../local-ic-vs-jcodemunch/artifacts/games/code-localization/rounds/round-001",
	}
	roundSpec.Artifacts.NextChallenger = &NextChallengerArtifact{
		Id:           "next-challenger-round-002",
		Kind:         "next_challenger",
		ArtifactName: "next_next_challenger_policy.round-002.py",
		Implements: Interface{
			Id: "iterative_context.selection_policy.v1",
		},
	}
	roundSpec.Agents.Optimizer = optimizer
	roundSpec.Optimization = &Optimization{
		Agent: *optimizer,
		ParentRound: ParentRound{
			Bundle: *roundSpec.Artifacts.ParentRoundBundle,
		},
		Target: NextChallengerTarget{
			Input:  *roundSpec.Artifacts.ChallengerPolicy,
			Output: *roundSpec.Artifacts.NextChallenger,
		},
		Evidence: NextChallengerEvidence{
			From: *roundSpec.Artifacts.ParentRoundBundle,
			Include: []NextChallengerEvidenceKind{
				NextChallengerEvidenceReportSummary,
				NextChallengerEvidenceRoundEvidence,
				NextChallengerEvidenceObjectiveResult,
				NextChallengerEvidenceChallengerPolicy,
			},
			Deny: []OptimizerDeniedEvidenceKind{
				OptimizerDeniedGoldLabels,
				OptimizerDeniedOracleFiles,
				OptimizerDeniedRawDatasetAnswers,
			},
		},
	}

	return roundSpec
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
