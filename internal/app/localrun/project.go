package localrun

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	artifact "github.com/becker63/searchbench-go/internal/adapters/artifact/fsbundle"
	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	"github.com/becker63/searchbench-go/internal/app/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

var errUnsupportedMode = errors.New("localrun: only evaluator_only mode is supported")

// Request configures one local fake manifest-driven run.
type Request struct {
	ManifestPath        string
	BundleRootOverride  string
	BundleID            string
	ReportID            domain.ReportID
	ParentRef           *score.ObjectiveEvidenceRef
	ParentScorePath     string
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

type resolvedPlan struct {
	manifestPath       string
	manifestDir        string
	experiment         config.Experiment
	comparePlan        compare.Plan
	resolvedInput      artifact.ResolvedComparisonInput
	objectivePath      string
	artifactRoot       domain.HostPath
	bundleCollection   domain.HostPath
	expectedBundlePath domain.HostPath
	bundleID           string
	reportID           domain.ReportID
	parentRef          *score.ObjectiveEvidenceRef
	parentScorePath    string
	createdAt          time.Time
	renderReport       bool
	timeout            time.Duration
}

type resolvedSystem struct {
	spec       domain.SystemSpec
	policyPath string
}

func normalizeRequest(request Request) Request {
	if request.Now == nil {
		request.Now = func() time.Time { return time.Now().UTC() }
	}
	return request
}

func resolvePlan(ctx context.Context, request Request) (resolvedPlan, error) {
	request = normalizeRequest(request)

	manifestPath, err := filepath.Abs(request.ManifestPath)
	if err != nil {
		return resolvedPlan{}, fmt.Errorf("resolve manifest path: %w", err)
	}

	experiment, err := config.ResolveFromPath(ctx, manifestPath)
	if err != nil {
		return resolvedPlan{}, err
	}
	if err := config.Validate(experiment); err != nil {
		return resolvedPlan{}, err
	}
	if experiment.Mode != config.ModeEvaluatorOnly {
		return resolvedPlan{}, fmt.Errorf("%w: %s", errUnsupportedMode, experiment.Mode)
	}

	manifestDir := filepath.Dir(manifestPath)
	objectivePath, err := resolveExistingManifestPath(manifestDir, experiment.Scoring.Objective)
	if err != nil {
		return resolvedPlan{}, fmt.Errorf("resolve objective path: %w", err)
	}
	bundleCollection, artifactRoot, err := resolveBundlePaths(manifestDir, experiment.OutputConfig.BundleRoot, request.BundleRootOverride)
	if err != nil {
		return resolvedPlan{}, fmt.Errorf("resolve bundle root: %w", err)
	}

	baseline, err := resolveSystem(manifestDir, experiment.Evaluator, experiment.Systems.Baseline)
	if err != nil {
		return resolvedPlan{}, fmt.Errorf("resolve baseline system: %w", err)
	}
	candidate, err := resolveSystem(manifestDir, experiment.Evaluator, experiment.Systems.Candidate)
	if err != nil {
		return resolvedPlan{}, fmt.Errorf("resolve candidate system: %w", err)
	}

	tasks := fakeTasks(manifestDir, experiment)
	comparePlan := compare.NewPlan(domain.NewPair(baseline.spec, candidate.spec), tasks)
	if err := comparePlan.Validate(); err != nil {
		return resolvedPlan{}, fmt.Errorf("validate local compare plan: %w", err)
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
	expectedBundlePath := domain.HostPath(filepath.Join(string(bundleCollection), bundleID))
	currentEvidenceRef := score.ObjectiveEvidenceRef{
		Name:       "current",
		BundlePath: string(expectedBundlePath),
		ScorePath:  filepath.Join(string(expectedBundlePath), "score.pkl"),
		ReportPath: filepath.Join(string(expectedBundlePath), "report.json"),
	}

	return resolvedPlan{
		manifestPath:       manifestPath,
		manifestDir:        manifestDir,
		experiment:         experiment,
		comparePlan:        comparePlan,
		objectivePath:      objectivePath,
		artifactRoot:       artifactRoot,
		bundleCollection:   bundleCollection,
		expectedBundlePath: expectedBundlePath,
		bundleID:           bundleID,
		reportID:           reportID,
		parentRef:          cloneEvidenceRef(request.ParentRef),
		parentScorePath:    request.ParentScorePath,
		createdAt:          now,
		renderReport:       renderReport,
		timeout:            timeoutFromSeconds(experiment.Evaluator.Bounds.TimeoutSeconds),
		resolvedInput: artifact.ResolvedComparisonInput{
			ManifestPath:   manifestPath,
			ExperimentName: experiment.Name,
			Mode:           experiment.Mode.String(),
			Dataset: artifact.DatasetConfig{
				Kind:     experiment.Dataset.Kind,
				Name:     experiment.Dataset.Name,
				Config:   experiment.Dataset.Config,
				Split:    experiment.Dataset.Split,
				MaxItems: experiment.Dataset.MaxItems,
			},
			Systems: comparePlan.ReportSpec().Systems,
			Tasks:   tasks,
			Parallelism: artifact.ParallelismConfig{
				Mode:       string(compare.ExecutionSequential),
				MaxWorkers: 1,
			},
			Evaluator: artifact.EvaluatorConfig{
				Model: artifact.EvaluatorModelConfig{
					Provider:        experiment.Evaluator.Model.Provider.String(),
					Name:            experiment.Evaluator.Model.Name,
					MaxOutputTokens: derefInt(experiment.Evaluator.Model.MaxOutputTokens),
				},
				Bounds: artifact.EvaluatorBoundsConfig{
					MaxModelTurns:  experiment.Evaluator.Bounds.MaxModelTurns,
					MaxToolCalls:   experiment.Evaluator.Bounds.MaxToolCalls,
					TimeoutSeconds: experiment.Evaluator.Bounds.TimeoutSeconds,
				},
				Retry: artifact.RetryPolicyConfig{
					MaxAttempts:                experiment.Evaluator.Retry.MaxAttempts,
					RetryOnModelError:          experiment.Evaluator.Retry.RetryOnModelError,
					RetryOnToolFailure:         experiment.Evaluator.Retry.RetryOnToolFailure,
					RetryOnFinalizationFailure: experiment.Evaluator.Retry.RetryOnFinalizationFailure,
					RetryOnInvalidPrediction:   experiment.Evaluator.Retry.RetryOnInvalidPrediction,
				},
			},
			Scoring: artifact.ScoringConfig{
				ObjectivePath: objectivePath,
				Evidence: artifact.EvidenceConfig{
					Current: currentEvidenceRef,
					Parent:  cloneEvidenceRef(request.ParentRef),
				},
			},
			Output: artifact.OutputConfig{
				BundleRoot:        filepath.ToSlash(string(bundleCollection)),
				BundleWriterRoot:  filepath.ToSlash(string(artifactRoot)),
				ReportFormat:      experiment.OutputConfig.ReportFormat.String(),
				RenderHumanReport: renderReport,
				ResolvedPolicyPath: artifact.ResolvedPolicyPath{
					Baseline:  filepath.ToSlash(baseline.policyPath),
					Candidate: filepath.ToSlash(candidate.policyPath),
				},
			},
			ReportOptions: artifact.ReportOptions{
				Format: experiment.OutputConfig.ReportFormat.String(),
			},
		},
	}, nil
}

func resolveSystem(manifestDir string, evaluator config.Evaluator, system config.System) (resolvedSystem, error) {
	backendKind, err := mapBackend(system.Backend)
	if err != nil {
		return resolvedSystem{}, err
	}

	spec := domain.SystemSpec{
		ID:      domain.SystemID(system.Id),
		Name:    system.Name,
		Backend: backendKind,
		Model: domain.ModelSpec{
			Provider: evaluator.Model.Provider.String(),
			Name:     evaluator.Model.Name,
		},
		PromptBundle: domain.PromptBundleRef{
			Name:    system.PromptBundle.Name,
			Version: derefString(system.PromptBundle.Version),
		},
		Runtime: domain.RuntimeConfig{
			MaxSteps:        resolvedMaxSteps(evaluator.Bounds.MaxModelTurns, system.Runtime.MaxSteps),
			MaxToolCalls:    evaluator.Bounds.MaxToolCalls,
			MaxOutputTokens: derefInt(evaluator.Model.MaxOutputTokens),
		},
	}

	out := resolvedSystem{spec: spec}
	if system.Policy == nil {
		return out, nil
	}

	policyPath, err := resolveExistingManifestPath(manifestDir, system.Policy.Path)
	if err != nil {
		return resolvedSystem{}, fmt.Errorf("resolve policy path: %w", err)
	}
	data, err := os.ReadFile(policyPath)
	if err != nil {
		return resolvedSystem{}, fmt.Errorf("read policy source: %w", err)
	}
	policy := domain.NewPythonPolicy(domain.PolicyID(system.Policy.Id), string(data), system.Policy.Entrypoint)
	out.spec.Policy = &policy
	out.policyPath = policyPath
	return out, nil
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

func resolveBundlePaths(manifestDir string, bundleRoot string, override string) (domain.HostPath, domain.HostPath, error) {
	collectionPath, err := resolveManifestPath(manifestDir, bundleRoot)
	if strings.TrimSpace(override) != "" {
		collectionPath, err = resolveOverridePath(override)
	}
	if err != nil {
		return "", "", err
	}
	if filepath.Base(collectionPath) == "runs" {
		return domain.HostPath(collectionPath), domain.HostPath(filepath.Dir(collectionPath)), nil
	}
	return domain.HostPath(filepath.Join(collectionPath, "runs")), domain.HostPath(collectionPath), nil
}

func resolveExistingManifestPath(baseDir string, path string) (string, error) {
	resolved, err := resolveManifestPath(baseDir, path)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(resolved); err != nil {
		return "", err
	}
	return resolved, nil
}

func resolveManifestPath(baseDir string, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(filepath.Join(baseDir, path))
}

func resolveOverridePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is required")
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	return filepath.Abs(path)
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

func resolvedMaxSteps(maxModelTurns int, fallback int) int {
	if maxModelTurns > 0 {
		return maxModelTurns
	}
	return fallback
}

func timeoutFromSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
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

func derefInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cloneEvidenceRef(ref *score.ObjectiveEvidenceRef) *score.ObjectiveEvidenceRef {
	if ref == nil {
		return nil
	}
	copyRef := *ref
	return &copyRef
}
