package experiment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	"github.com/becker63/searchbench-go/internal/app/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

var ErrUnsupportedMode = errors.New("experiment: only evaluator_only mode is supported")

// Resolve loads one Pkl manifest through the config adapter and projects it
// into the canonical resolved experiment plan.
func Resolve(ctx context.Context, request Request) (ResolvedExperiment, error) {
	request = normalizeRequest(request)

	manifestPath, err := filepath.Abs(request.ManifestPath)
	if err != nil {
		return ResolvedExperiment{}, fmt.Errorf("resolve manifest path: %w", err)
	}

	cfg, err := config.ResolveFromPath(ctx, manifestPath)
	if err != nil {
		return ResolvedExperiment{}, err
	}
	if err := config.Validate(cfg); err != nil {
		return ResolvedExperiment{}, err
	}
	if cfg.Mode != config.ModeEvaluatorOnly {
		return ResolvedExperiment{}, fmt.Errorf("%w: %s", ErrUnsupportedMode, cfg.Mode)
	}

	manifestDir := filepath.Dir(manifestPath)
	objectivePath, err := resolveExistingManifestPath(manifestDir, cfg.Scoring.Objective)
	if err != nil {
		return ResolvedExperiment{}, fmt.Errorf("resolve objective path: %w", err)
	}
	bundleCollection, bundleWriterRoot, err := resolveBundlePaths(manifestDir, cfg.OutputConfig.BundleRoot, request.BundleRootOverride)
	if err != nil {
		return ResolvedExperiment{}, fmt.Errorf("resolve bundle root: %w", err)
	}

	baseline, err := resolveSystem(manifestDir, cfg.Evaluator, cfg.Systems.Baseline)
	if err != nil {
		return ResolvedExperiment{}, fmt.Errorf("resolve baseline system: %w", err)
	}
	candidate, err := resolveSystem(manifestDir, cfg.Evaluator, cfg.Systems.Candidate)
	if err != nil {
		return ResolvedExperiment{}, fmt.Errorf("resolve candidate system: %w", err)
	}

	tasks := fakeTasks(manifestDir, cfg)
	systems := domain.NewPair(baseline.spec, candidate.spec)
	comparePlan := compare.NewPlan(systems, tasks)
	if err := comparePlan.Validate(); err != nil {
		return ResolvedExperiment{}, fmt.Errorf("validate compare plan: %w", err)
	}

	now := request.Now().UTC()
	bundleID := request.BundleID
	if bundleID == "" {
		bundleID = defaultBundleID(cfg.Name, now)
	}
	reportID := request.ReportID
	if reportID.Empty() {
		reportID = defaultReportID(bundleID)
	}

	expectedBundlePath := domain.HostPath(filepath.Join(string(bundleCollection), bundleID))
	renderHumanReport := !request.DisableRenderReport && cfg.OutputConfig.ReportFormat != config.ReportFormatJSON

	return ResolvedExperiment{
		ManifestPath:   manifestPath,
		ExperimentName: cfg.Name,
		Mode:           cfg.Mode.String(),
		Dataset: DatasetConfig{
			Kind:     cfg.Dataset.Kind,
			Name:     cfg.Dataset.Name,
			Config:   cfg.Dataset.Config,
			Split:    cfg.Dataset.Split,
			MaxItems: cfg.Dataset.MaxItems,
		},
		Systems:     systems,
		Tasks:       tasks,
		Parallelism: compare.DefaultParallelism(),
		Evaluator: EvaluatorConfig{
			Model: EvaluatorModelConfig{
				Provider:        cfg.Evaluator.Model.Provider.String(),
				Name:            cfg.Evaluator.Model.Name,
				MaxOutputTokens: derefInt(cfg.Evaluator.Model.MaxOutputTokens),
			},
			Bounds: EvaluatorBoundsConfig{
				MaxModelTurns:  cfg.Evaluator.Bounds.MaxModelTurns,
				MaxToolCalls:   cfg.Evaluator.Bounds.MaxToolCalls,
				TimeoutSeconds: cfg.Evaluator.Bounds.TimeoutSeconds,
			},
			Retry: RetryPolicyConfig{
				MaxAttempts:                cfg.Evaluator.Retry.MaxAttempts,
				RetryOnModelError:          cfg.Evaluator.Retry.RetryOnModelError,
				RetryOnToolFailure:         cfg.Evaluator.Retry.RetryOnToolFailure,
				RetryOnFinalizationFailure: cfg.Evaluator.Retry.RetryOnFinalizationFailure,
				RetryOnInvalidPrediction:   cfg.Evaluator.Retry.RetryOnInvalidPrediction,
			},
		},
		Scoring: ScoringConfig{
			ObjectivePath: objectivePath,
			CurrentEvidence: score.ObjectiveEvidenceRef{
				Name:       "current",
				BundlePath: string(expectedBundlePath),
				ScorePath:  filepath.Join(string(expectedBundlePath), "score.pkl"),
				ReportPath: filepath.Join(string(expectedBundlePath), "report.json"),
			},
			ParentEvidence:  cloneEvidenceRef(request.ParentRef),
			ParentScorePath: request.ParentScorePath,
		},
		Output: OutputConfig{
			BundleCollectionPath: bundleCollection,
			BundleWriterRoot:     bundleWriterRoot,
			ExpectedBundlePath:   expectedBundlePath,
			ReportFormat:         cfg.OutputConfig.ReportFormat.String(),
			RenderHumanReport:    renderHumanReport,
			ResolvedPolicyPaths: ResolvedPolicyPaths{
				Baseline:  filepath.ToSlash(baseline.policyPath),
				Candidate: filepath.ToSlash(candidate.policyPath),
			},
		},
		Report: ReportConfig{
			Format: cfg.OutputConfig.ReportFormat.String(),
		},
		Bundle: BundleConfig{
			ID: bundleID,
		},
		ReportID:  reportID,
		CreatedAt: now,
	}, nil
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

func resolveSystem(manifestDir string, evaluator config.Evaluator, system config.System) (resolvedSystem, error) {
	backendKind, err := mapBackend(system.Backend)
	if err != nil {
		return resolvedSystem{}, err
	}

	out := resolvedSystem{
		spec: domain.SystemSpec{
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
		},
	}
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

func fakeTasks(manifestDir string, cfg config.Experiment) domain.NonEmpty[domain.TaskSpec] {
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
			Title: fmt.Sprintf("Fake %s/%s task", cfg.Dataset.Config, cfg.Dataset.Split),
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

func defaultBundleID(name string, now time.Time) string {
	return sanitizeBundleID(name) + "-" + strings.ToLower(now.UTC().Format("20060102t150405z"))
}

func defaultReportID(bundleID string) domain.ReportID {
	return domain.ReportID("report-" + bundleID)
}

func sanitizeBundleID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "experiment"
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
