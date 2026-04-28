package localrun

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/becker63/searchbench-go/internal/artifact"
	"github.com/becker63/searchbench-go/internal/compare"
	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/score"
	"github.com/becker63/searchbench-go/internal/surface/config"
)

var errUnsupportedMode = errors.New("localrun: only evaluator_only mode is supported")

// Request configures one local fake manifest-driven run.
type Request struct {
	ManifestPath        string
	BundleRootOverride  string
	BundleID            string
	ReportID            domain.ReportID
	Now                 func() time.Time
	PklCommand          []string
	DisableRenderReport bool
}

// Result is the completed local fake manifest-driven run.
type Result struct {
	ManifestPath    string
	Bundle          artifact.BundleRef
	ReportID        domain.ReportID
	CandidateReport report.CandidateReport
	ScoreEvidence   score.ScoreEvidenceDocument
	ObjectiveResult *score.ObjectiveResult
}

type projectedRun struct {
	manifestPath       string
	manifestDir        string
	experiment         config.Experiment
	plan               compare.Plan
	resolvedInput      artifact.ResolvedComparisonInput
	objectivePath      string
	artifactRoot       domain.HostPath
	bundleCollection   domain.HostPath
	expectedBundlePath domain.HostPath
	bundleID           string
	reportID           domain.ReportID
	createdAt          time.Time
	renderReport       bool
}

func normalizeRequest(request Request) Request {
	if request.Now == nil {
		request.Now = func() time.Time { return time.Now().UTC() }
	}
	return request
}

func projectFakeRun(manifestPath string, experiment config.Experiment, request Request) (projectedRun, error) {
	if experiment.Mode != config.ModeEvaluatorOnly {
		return projectedRun{}, fmt.Errorf("%w: %s", errUnsupportedMode, experiment.Mode)
	}

	manifestDir := filepath.Dir(manifestPath)
	objectivePath, err := resolvePath(manifestDir, experiment.Scoring.Objective)
	if err != nil {
		return projectedRun{}, fmt.Errorf("resolve objective path: %w", err)
	}

	bundleCollection, artifactRoot, err := resolveBundlePaths(manifestDir, experiment.OutputConfig.BundleRoot, request.BundleRootOverride)
	if err != nil {
		return projectedRun{}, fmt.Errorf("resolve bundle root: %w", err)
	}

	systems, err := projectSystems(manifestDir, experiment)
	if err != nil {
		return projectedRun{}, err
	}
	tasks := fakeTasks(manifestDir, experiment)
	plan := compare.NewPlan(systems, tasks)
	if err := plan.Validate(); err != nil {
		return projectedRun{}, fmt.Errorf("validate fake plan: %w", err)
	}

	now := request.Now().UTC()
	bundleID := request.BundleID
	if bundleID == "" {
		bundleID = defaultBundleID(experiment.Name, now)
	}
	reportID := request.ReportID
	if reportID.Empty() {
		reportID = defaultReportID(bundleID)
	}

	renderReport := !request.DisableRenderReport && experiment.OutputConfig.ReportFormat != config.ReportFormatJSON

	return projectedRun{
		manifestPath:       manifestPath,
		manifestDir:        manifestDir,
		experiment:         experiment,
		plan:               plan,
		objectivePath:      objectivePath,
		artifactRoot:       artifactRoot,
		bundleCollection:   bundleCollection,
		expectedBundlePath: domain.HostPath(filepath.Join(string(bundleCollection), bundleID)),
		bundleID:           bundleID,
		reportID:           reportID,
		createdAt:          now,
		renderReport:       renderReport,
		resolvedInput: artifact.ResolvedComparisonInput{
			Systems: plan.ReportSpec().Systems,
			Tasks:   tasks,
			Parallelism: artifact.ParallelismConfig{
				Mode:       string(compare.ExecutionSequential),
				MaxWorkers: 1,
			},
			ScoringProfile: filepath.ToSlash(experiment.Scoring.Objective),
			ReportOptions: artifact.ReportOptions{
				Format: experiment.OutputConfig.ReportFormat.String(),
			},
		},
	}, nil
}

func projectSystems(manifestDir string, experiment config.Experiment) (domain.Pair[domain.SystemSpec], error) {
	model := domain.ModelSpec{
		Provider: experiment.Evaluator.Model.Provider.String(),
		Name:     experiment.Evaluator.Model.Name,
	}
	outputTokens := 0
	if experiment.Evaluator.Model.MaxOutputTokens != nil {
		outputTokens = *experiment.Evaluator.Model.MaxOutputTokens
	}

	project := func(system config.System) (domain.SystemSpec, error) {
		backendKind, err := mapBackend(system.Backend)
		if err != nil {
			return domain.SystemSpec{}, err
		}

		spec := domain.SystemSpec{
			ID:      domain.SystemID(system.Id),
			Name:    system.Name,
			Backend: backendKind,
			Model:   model,
			PromptBundle: domain.PromptBundleRef{
				Name:    system.PromptBundle.Name,
				Version: derefString(system.PromptBundle.Version),
			},
			Runtime: domain.RuntimeConfig{
				MaxSteps:        system.Runtime.MaxSteps,
				MaxToolCalls:    experiment.Evaluator.Bounds.MaxToolCalls,
				MaxOutputTokens: outputTokens,
			},
		}
		if system.Policy != nil {
			policy, err := loadPolicy(manifestDir, *system.Policy)
			if err != nil {
				return domain.SystemSpec{}, err
			}
			spec.Policy = &policy
		}
		return spec, nil
	}

	baseline, err := project(experiment.Systems.Baseline)
	if err != nil {
		return domain.Pair[domain.SystemSpec]{}, fmt.Errorf("project baseline system: %w", err)
	}
	candidate, err := project(experiment.Systems.Candidate)
	if err != nil {
		return domain.Pair[domain.SystemSpec]{}, fmt.Errorf("project candidate system: %w", err)
	}
	return domain.NewPair(baseline, candidate), nil
}

func fakeTasks(manifestDir string, experiment config.Experiment) domain.NonEmpty[domain.TaskSpec] {
	repoPath := domain.HostPath(filepath.Join(manifestDir, "fake-repo"))
	task := domain.TaskSpec{
		ID:        domain.TaskID("local-fake-task-1"),
		Benchmark: domain.BenchmarkLCA,
		Repo: domain.RepoSnapshot{
			Name: domain.RepoName("searchbench/local-fake"),
			SHA:  domain.RepoSHA("0000000"),
			Path: repoPath,
		},
		Input: domain.TaskInput{
			Title: fmt.Sprintf("Fake %s/%s task", experiment.Dataset.Config, experiment.Dataset.Split),
			Body:  "This deterministic local task exists only to prove manifest-driven composition.",
		},
		Oracle: domain.TaskOracle{
			GoldFiles: []domain.RepoRelPath{"src/search_target.go"},
		},
	}
	return domain.NewNonEmpty(task)
}

func loadPolicy(manifestDir string, policy config.Policy) (domain.PolicyArtifact, error) {
	path, err := resolvePath(manifestDir, policy.Path)
	if err != nil {
		return domain.PolicyArtifact{}, fmt.Errorf("resolve policy path: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.PolicyArtifact{}, fmt.Errorf("read policy source: %w", err)
	}
	return domain.NewPythonPolicy(domain.PolicyID(policy.Id), string(data), policy.Entrypoint), nil
}

func resolveBundlePaths(manifestDir string, bundleRoot string, override string) (domain.HostPath, domain.HostPath, error) {
	collectionPath, err := resolvePath(manifestDir, firstNonEmptyString(override, bundleRoot))
	if err != nil {
		return "", "", err
	}
	if filepath.Base(collectionPath) == "runs" {
		return domain.HostPath(collectionPath), domain.HostPath(filepath.Dir(collectionPath)), nil
	}
	return domain.HostPath(filepath.Join(collectionPath, "runs")), domain.HostPath(collectionPath), nil
}

func resolvePath(baseDir string, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join(baseDir, path))
}

func mapBackend(backend config.Backend) (domain.BackendKind, error) {
	switch backend {
	case config.BackendIterativeContext:
		return domain.BackendIterativeContext, nil
	case config.BackendJCodeMunch:
		return domain.BackendJCodeMunch, nil
	case config.BackendFake:
		return domain.BackendFake, nil
	default:
		return "", fmt.Errorf("unsupported backend %q", backend)
	}
}

func defaultBundleID(name string, now time.Time) string {
	return sanitizeBundleID(name) + "-" + strings.ToLower(now.UTC().Format("20060102t150405z"))
}

func defaultReportID(bundleID string) domain.ReportID {
	return domain.ReportID("report-" + bundleID)
}

func sanitizeBundleID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "localrun"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
