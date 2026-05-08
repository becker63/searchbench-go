package config

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestValidEvaluationConfigValidates(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	if err := Validate(experiment); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidOptimizationConfigValidates(t *testing.T) {
	t.Parallel()

	experiment := sampleOptimizationExperiment()
	if experiment.Evaluation != nil {
		t.Fatalf("sampleOptimizationExperiment().Evaluation = %#v, want nil", experiment.Evaluation)
	}
	if err := Validate(experiment); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestOptimizationModeDoesNotRequireEvaluationOrScoring(t *testing.T) {
	t.Parallel()

	experiment := sampleOptimizationExperiment()
	experiment.Evaluation = nil

	if err := Validate(experiment); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestEvaluationModeRequiresEvaluation(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Evaluation = nil

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingEvaluation.Error()) {
		t.Fatalf("Validate() error = %v, want missing evaluation error", err)
	}
}

func TestEvaluationModeRequiresEvaluator(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Agents.Evaluator = nil

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingEvaluator.Error()) {
		t.Fatalf("Validate() error = %v, want missing evaluator error", err)
	}
}

func TestOptimizationModeRequiresOptimization(t *testing.T) {
	t.Parallel()

	experiment := sampleOptimizationExperiment()
	experiment.Optimization = nil

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingOptimization.Error()) {
		t.Fatalf("Validate() error = %v, want missing optimization error", err)
	}
}

func TestOptimizationModeRequiresOptimizer(t *testing.T) {
	t.Parallel()

	experiment := sampleOptimizationExperiment()
	experiment.Agents.Optimizer = nil

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingOptimizer.Error()) {
		t.Fatalf("Validate() error = %v, want missing optimizer error", err)
	}
}

func TestExperimentConfigNoLongerHasWriterOrOutputConfig(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(Experiment{})
	if _, ok := typ.FieldByName("Writer"); ok {
		t.Fatal("Experiment unexpectedly has Writer field")
	}
	if _, ok := typ.FieldByName("OutputConfig"); ok {
		t.Fatal("Experiment unexpectedly has OutputConfig field")
	}
}

func TestRejectsMissingDatasetConfig(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Dataset.Config = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingDatasetConfig.Error()) {
		t.Fatalf("Validate() error = %v, want missing dataset config error", err)
	}
}

func TestRejectsMissingDatasetSplit(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Dataset.Split = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingDatasetSplit.Error()) {
		t.Fatalf("Validate() error = %v, want missing dataset split error", err)
	}
}

func TestRejectsMissingBaselineSystemID(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Systems.Baseline.Id = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingBaselineSystemID.Error()) {
		t.Fatalf("Validate() error = %v, want missing baseline id error", err)
	}
}

func TestRejectsMissingIterativeContextSystemID(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Systems.IterativeContext.Id = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingIterativeContextSystemID.Error()) {
		t.Fatalf("Validate() error = %v, want missing iterative context id error", err)
	}
}

func TestRejectsMissingAgentModelProvider(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Agents.Evaluator.Model.Provider = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingAgentModelProvider.Error()) {
		t.Fatalf("Validate() error = %v, want missing provider error", err)
	}
}

func TestRejectsMissingAgentModelName(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Agents.Evaluator.Model.Name = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingAgentModelName.Error()) {
		t.Fatalf("Validate() error = %v, want missing model name error", err)
	}
}

func TestRejectsMissingScoringObjectivePath(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Evaluation.Scoring.Objective = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingScoringObjectivePath.Error()) {
		t.Fatalf("Validate() error = %v, want missing scoring objective error", err)
	}
}

func TestRejectsEvaluationAgentMismatch(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Evaluation.Agent.Model.Name = "other"

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrEvaluationAgentMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want evaluation agent mismatch", err)
	}
}

func TestRejectsEvaluationCandidateSystemMismatch(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Evaluation.Candidate.System.Id = "other"

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrEvaluationCandidateSystemMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want candidate system mismatch", err)
	}
}

func TestRejectsSelectionPolicyArtifactMismatch(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Evaluation.Candidate.Uses.SelectionPolicy.Id = "other"

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrSelectionPolicyArtifactMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want selection policy mismatch", err)
	}
}

func TestRejectsSelectionPolicyInterfaceMismatch(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Artifacts.CandidatePolicyRound001.Implements.Id = "other"
	experiment.Evaluation.Candidate.Uses.SelectionPolicy.Implements.Id = "other"

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrSelectionPolicyInterfaceMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want selection policy interface mismatch", err)
	}
}

func TestRejectsAbsolutePolicyArtifactPath(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Artifacts.CandidatePolicyRound001.Path = "/tmp/policy.py"

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrPolicyArtifactPathMustBeRelative.Error()) {
		t.Fatalf("Validate() error = %v, want relative policy path error", err)
	}
}

func TestRejectsProposalArtifactNameWithParentTraversal(t *testing.T) {
	t.Parallel()

	experiment := sampleOptimizationExperiment()
	experiment.Artifacts.CandidatePolicyRound002.ArtifactName = "../candidate.py"

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrPolicyProposalArtifactNameInvalid.Error()) {
		t.Fatalf("Validate() error = %v, want invalid artifact name error", err)
	}
}

func TestRejectsOptimizationTargetOutputMismatch(t *testing.T) {
	t.Parallel()

	experiment := sampleOptimizationExperiment()
	experiment.Optimization.Target.Output.Id = "other"

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrOptimizationTargetOutputMismatch.Error()) {
		t.Fatalf("Validate() error = %v, want optimization target output mismatch", err)
	}
}

func TestRejectsEmptyToolAllowEntry(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Agents.Evaluator.Tools.Allow = append(experiment.Agents.Evaluator.Tools.Allow, " ")

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrToolAllowEntryEmpty.Error()) {
		t.Fatalf("Validate() error = %v, want empty allow entry error", err)
	}
}

func TestRejectsDuplicateToolAllowEntry(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Agents.Evaluator.Tools.Allow = append(experiment.Agents.Evaluator.Tools.Allow, experiment.Agents.Evaluator.Tools.Allow[0])

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrToolAllowDuplicate.Error()) {
		t.Fatalf("Validate() error = %v, want duplicate allow entry error", err)
	}
}

func TestRejectsToolAllowDenyOverlap(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Agents.Evaluator.Tools.Deny = append(experiment.Agents.Evaluator.Tools.Deny, experiment.Agents.Evaluator.Tools.Allow[0])

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrToolPolicyOverlap.Error()) {
		t.Fatalf("Validate() error = %v, want tool overlap error", err)
	}
}

func TestRejectsOversizedSystemPrompt(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	oversized := strings.Repeat("x", maxSystemPromptBytes+1)
	experiment.Agents.Evaluator.SystemPrompt = &oversized
	experiment.Evaluation.Agent.SystemPrompt = &oversized

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrSystemPromptTooLarge.Error()) {
		t.Fatalf("Validate() error = %v, want oversized prompt error", err)
	}
}

func TestRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluationExperiment()
	experiment.Mode = RunMode("mystery_mode")

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrUnsupportedMode.Error()) {
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
		filepath.Join(repoRoot, "internal", "pure", "run"),
		filepath.Join(repoRoot, "internal", "pure", "score"),
		filepath.Join(repoRoot, "internal", "pure", "report"),
		filepath.Join(repoRoot, "internal", "pure", "codegraph"),
		filepath.Join(repoRoot, "internal", "pure", "prompts"),
		filepath.Join(repoRoot, "internal", "pure", "usage"),
	}

	for _, dir := range dirs {
		fs := token.NewFileSet()
		pkgs, err := parser.ParseDir(fs, dir, func(info os.FileInfo) bool {
			name := info.Name()
			return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
		}, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parser.ParseDir(%q) error = %v", dir, err)
		}
		for _, pkg := range pkgs {
			for _, file := range pkg.Files {
				for _, imp := range file.Imports {
					path := strings.Trim(imp.Path.Value, `"`)
					if strings.Contains(path, "github.com/apple/pkl-go") {
						t.Fatalf("pure package import %q leaked pkl-go", path)
					}
				}
			}
		}
	}
}

func sampleEvaluationExperiment() Experiment {
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

	experiment := Experiment{
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
		Systems: Systems{
			Baseline: System{
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
			IterativeContext: System{
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
			CandidatePolicyRound001: &PolicyArtifact{
				Id:   "candidate-policy-round-001",
				Kind: "policy",
				Path: "policies/candidate_policy.py",
				Implements: Interface{
					Id: "iterative_context.selection_policy.v1",
				},
			},
		},
		Agents: Agents{
			Evaluator: evaluator,
		},
	}

	experiment.Evaluation = &Evaluation{
		Agent: *evaluator,
		Baseline: EvaluationSystemBinding{
			System: experiment.Systems.Baseline,
		},
		Candidate: CandidateEvaluationBinding{
			System: experiment.Systems.IterativeContext,
			Uses: CandidateUses{
				SelectionPolicy: *experiment.Artifacts.CandidatePolicyRound001,
			},
		},
		Scoring: Scoring{
			Objective: "scoring/localization-objective.pkl",
		},
		Report: Report{
			Formats: []ReportFormat{ReportFormatJSON, ReportFormatText},
		},
	}

	return experiment
}

func sampleOptimizationExperiment() Experiment {
	experiment := sampleEvaluationExperiment()
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
				"read_score_evidence",
				"read_objective_result",
				"read_report_summary",
				"read_artifact",
				"write_artifact_proposal",
			},
			Deny: []string{"run_evaluator", "shell", "go_test", "write_manifest", "network"},
		},
		SystemPrompt: &prompt,
	}

	experiment.Name = "optimize-ic-round-002"
	experiment.Mode = ModeOptimization
	experiment.Evaluation = nil
	experiment.Artifacts.ParentEvaluationRound001 = &CompletedEvaluationBundleArtifact{
		Id:   "local-ic-vs-jcodemunch-round-001",
		Kind: "completed_evaluation_bundle",
		Path: "../local-ic-vs-jcodemunch/artifacts/runs/example-round-001",
	}
	experiment.Artifacts.CandidatePolicyRound002 = &PolicyProposalArtifact{
		Id:           "candidate-policy-round-002",
		Kind:         "policy_proposal",
		ArtifactName: "candidate_policy.round-002.py",
		Implements: Interface{
			Id: "iterative_context.selection_policy.v1",
		},
	}
	experiment.Agents.Optimizer = optimizer
	experiment.Optimization = &Optimization{
		Agent: *optimizer,
		ParentRun: ParentRun{
			Bundle: *experiment.Artifacts.ParentEvaluationRound001,
		},
		Target: OptimizationTarget{
			Input:  *experiment.Artifacts.CandidatePolicyRound001,
			Output: *experiment.Artifacts.CandidatePolicyRound002,
		},
		Evidence: OptimizationEvidence{
			From: *experiment.Artifacts.ParentEvaluationRound001,
			Include: []OptimizerEvidenceKind{
				OptimizerEvidenceReportSummary,
				OptimizerEvidenceScoreEvidence,
				OptimizerEvidenceObjectiveResult,
				OptimizerEvidenceCandidatePolicy,
			},
			Deny: []OptimizerDeniedEvidenceKind{
				OptimizerDeniedGoldLabels,
				OptimizerDeniedOracleFiles,
				OptimizerDeniedRawDatasetAnswers,
			},
		},
	}

	return experiment
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
