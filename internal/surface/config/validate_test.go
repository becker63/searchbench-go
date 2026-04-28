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

func TestValidEvaluatorOnlyConfigValidates(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	if err := Validate(experiment); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestEvaluatorOnlyRejectsEnabledWriter(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Writer = &Writer{
		Enabled: true,
		Model:   sampleModel(),
	}

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrWriterNotAllowed.Error()) {
		t.Fatalf("Validate() error = %v, want writer not allowed error", err)
	}
}

func TestWriterOptimizationRequiresWriter(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Mode = ModeWriterOptimization
	experiment.Writer = nil

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrWriterRequired.Error()) {
		t.Fatalf("Validate() error = %v, want writer required error", err)
	}
}

func TestEvaluatorConfigHasNoPipelineField(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(Evaluator{})
	if _, ok := typ.FieldByName("Pipeline"); ok {
		t.Fatal("Evaluator unexpectedly has Pipeline field")
	}
}

func TestWriterPipelineAcceptsNamedStepsWithArgv(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Mode = ModeWriterOptimization
	experiment.Writer = sampleWriter()

	if err := Validate(experiment); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestWriterPipelineRejectsEmptyStepName(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Mode = ModeWriterOptimization
	experiment.Writer = sampleWriter()
	experiment.Writer.Pipeline.Steps[0].Name = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrEmptyPipelineStepName.Error()) {
		t.Fatalf("Validate() error = %v, want empty step name error", err)
	}
}

func TestWriterPipelineRejectsEmptyArgv(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Mode = ModeWriterOptimization
	experiment.Writer = sampleWriter()
	experiment.Writer.Pipeline.Steps[0].Argv = nil

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrEmptyPipelineStepArgv.Error()) {
		t.Fatalf("Validate() error = %v, want empty step argv error", err)
	}
}

func TestRejectsMissingDatasetConfig(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Dataset.Config = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingDatasetConfig.Error()) {
		t.Fatalf("Validate() error = %v, want missing dataset config error", err)
	}
}

func TestRejectsMissingDatasetSplit(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Dataset.Split = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingDatasetSplit.Error()) {
		t.Fatalf("Validate() error = %v, want missing dataset split error", err)
	}
}

func TestRejectsMissingBaselineSystemID(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Systems.Baseline.Id = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingBaselineSystemID.Error()) {
		t.Fatalf("Validate() error = %v, want missing baseline id error", err)
	}
}

func TestRejectsMissingCandidateSystemID(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Systems.Candidate.Id = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingCandidateSystemID.Error()) {
		t.Fatalf("Validate() error = %v, want missing candidate id error", err)
	}
}

func TestRejectsMissingEvaluatorModelProvider(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Evaluator.Model.Provider = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingEvaluatorProvider.Error()) {
		t.Fatalf("Validate() error = %v, want missing evaluator provider error", err)
	}
}

func TestRejectsMissingEvaluatorModelName(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Evaluator.Model.Name = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingEvaluatorModelName.Error()) {
		t.Fatalf("Validate() error = %v, want missing evaluator model name error", err)
	}
}

func TestRejectsMissingScoringObjectivePath(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
	experiment.Scoring.Objective = ""

	if err := Validate(experiment); err == nil || !strings.Contains(err.Error(), ErrMissingScoringObjectivePath.Error()) {
		t.Fatalf("Validate() error = %v, want missing scoring objective error", err)
	}
}

func TestRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	experiment := sampleEvaluatorOnlyExperiment()
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

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	dirs := []string{
		filepath.Join(repoRoot, "internal", "domain"),
		filepath.Join(repoRoot, "internal", "run"),
		filepath.Join(repoRoot, "internal", "score"),
		filepath.Join(repoRoot, "internal", "report"),
		filepath.Join(repoRoot, "internal", "compare"),
		filepath.Join(repoRoot, "internal", "executor"),
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

func sampleEvaluatorOnlyExperiment() Experiment {
	return Experiment{
		Name: "local-ic-vs-jcodemunch-lca-dev",
		Mode: ModeEvaluatorOnly,
		Dataset: Dataset{
			Kind:     "lca",
			Name:     "JetBrains-Research/lca-bug-localization",
			Config:   "py",
			Split:    "dev",
			MaxItems: intPtr(5),
		},
		Systems: Systems{
			Baseline: System{
				Id:      "jcodemunch-baseline",
				Name:    "jCodeMunch baseline",
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
			Candidate: System{
				Id:      "iterative-context-candidate",
				Name:    "Iterative Context candidate",
				Backend: BackendIterativeContext,
				PromptBundle: PromptBundle{
					Name:    "default",
					Version: stringPtr("dev"),
				},
				Runtime: Runtime{
					MaxSteps:       16,
					TimeoutSeconds: 300,
				},
				Policy: &Policy{
					Id:         "candidate-policy-dev",
					Language:   "python",
					Entrypoint: "score",
					Path:       "policies/candidate_policy.py",
				},
			},
		},
		Evaluator: Evaluator{
			Model: sampleModel(),
			Bounds: AgentBounds{
				MaxModelTurns:  8,
				MaxToolCalls:   24,
				TimeoutSeconds: 300,
			},
			Retry: RetryPolicy{
				MaxAttempts:                2,
				RetryOnModelError:          true,
				RetryOnToolFailure:         false,
				RetryOnFinalizationFailure: true,
				RetryOnInvalidPrediction:   true,
			},
		},
		Scoring: Scoring{
			Objective: "scoring/localization-objective.pkl",
		},
		OutputConfig: Output{
			ReportFormat: ReportFormatPretty,
			BundleRoot:   "artifacts/runs",
			Traces: Tracing{
				Enabled: false,
			},
		},
	}
}

func sampleWriter() *Writer {
	return &Writer{
		Enabled:     true,
		Model:       sampleModel(),
		MaxAttempts: 3,
		Pipeline: &PipelineProfile{
			Name: "go-and-python-policy",
			Steps: []PipelineStep{
				{
					Name: "go_test",
					Argv: []string{"go", "test", "./..."},
				},
			},
		},
	}
}

func sampleModel() Model {
	return Model{
		Provider:        ProviderOpenRouter,
		Name:            "moonshotai/kimi-k2",
		MaxOutputTokens: intPtr(2000),
	}
}

func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
