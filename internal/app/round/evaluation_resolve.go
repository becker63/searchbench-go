package round

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	datasetfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/ports/dataset"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// defaultMatchSource is the deterministic local match source used when the
// caller does not inject a real dataset adapter. It is wired through the
// MatchSource port so a real adapter can be substituted without touching
// resolveEvaluation.
var defaultMatchSource dataset.MatchSource = datasetfake.NewMatchSource()

// selectionPolicyV1DefaultSymbol is the runtime callable used when adapting
// the manifest-level iterative_context.selection_policy.v1 interface into the
// existing domain.Policy shape. It is not the SearchBench objective score.
const selectionPolicyV1DefaultSymbol = "score"

var ErrUnsupportedMode = errors.New("evaluation: only evaluation mode is supported")

// resolveEvaluation loads one Pkl manifest through the config adapter and
// projects it into the canonical resolved round plan.
func resolveEvaluation(ctx context.Context, request evaluationResolveRequest) (Plan, error) {
	request = normalizeEvaluationRequest(request)

	manifestPath, err := filepath.Abs(request.ManifestPath)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve manifest path: %w", err)
	}

	cfg, err := config.ResolveFromPath(ctx, manifestPath)
	if err != nil {
		return Plan{}, err
	}
	if err := config.Validate(cfg); err != nil {
		return Plan{}, err
	}
	if cfg.Mode != config.ModeEvaluation {
		return Plan{}, fmt.Errorf("%w: %s", ErrUnsupportedMode, cfg.Mode)
	}
	if cfg.Evaluation == nil || cfg.Agents.Evaluator == nil {
		return Plan{}, fmt.Errorf("%w: incomplete evaluation manifest", ErrUnsupportedMode)
	}

	manifestDir := filepath.Dir(manifestPath)
	evaluation := cfg.Evaluation
	evaluator := cfg.Agents.Evaluator

	objectivePath, err := resolveExistingManifestPath(manifestDir, evaluation.Scoring.Objective)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve objective path: %w", err)
	}
	_, bundleWriterRoot, err := resolveBundlePaths(manifestDir, request.BundleRootOverride)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve bundle root: %w", err)
	}

	incumbent, err := resolveSystem(manifestDir, *evaluator, evaluation.Incumbent.System, nil)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve incumbent policy: %w", err)
	}
	challenger, err := resolveSystem(
		manifestDir,
		*evaluator,
		evaluation.Challenger.System,
		&evaluation.Challenger.Uses.SelectionPolicy,
	)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve challenger policy: %w", err)
	}

	tasks, err := defaultMatchSource.Matches(ctx, dataset.Request{
		ManifestDir: manifestDir,
		Kind:        cfg.Dataset.Kind,
		Name:        cfg.Dataset.Name,
		Config:      cfg.Dataset.Config,
		Split:       cfg.Dataset.Split,
		MaxItems:    cfg.Dataset.MaxItems,
	})
	if err != nil {
		return Plan{}, fmt.Errorf("resolve matches: %w", err)
	}
	systems := domain.NewPair(incumbent.spec, challenger.spec)
	comparePlan := compare.NewPlan(systems, tasks)
	if err := comparePlan.Validate(); err != nil {
		return Plan{}, fmt.Errorf("validate compare plan: %w", err)
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

	gameID := cfg.Game.Id
	if strings.TrimSpace(gameID) == "" {
		gameID = "code-localization"
	}
	bundleCollection := domain.HostPath(filepath.Join(string(bundleWriterRoot), "games", gameID, "rounds"))
	expectedBundlePath := domain.HostPath(filepath.Join(string(bundleCollection), bundleID))
	reportFormats := stringifyReportFormats(evaluation.Report.Formats)
	renderHumanReport := containsReportFormat(reportFormats, config.ReportFormatText.String())

	return Plan{
		ManifestPath: manifestPath,
		RoundName:    cfg.Name,
		Mode:         cfg.Mode.String(),
		Game: GameConfig{
			ID:   cfg.Game.Id,
			Kind: cfg.Game.Kind,
		},
		Round: RoundConfig{
			ID: bundleID,
		},
		Dataset: DatasetConfig{
			Kind:     cfg.Dataset.Kind,
			Name:     cfg.Dataset.Name,
			Config:   cfg.Dataset.Config,
			Split:    cfg.Dataset.Split,
			MaxItems: cfg.Dataset.MaxItems,
		},
		Policies:    systems,
		Matches:     tasks,
		Parallelism: compare.DefaultParallelism(),
		Evaluator: EvaluatorConfig{
			Model: EvaluatorModelConfig{
				Provider:        evaluator.Model.Provider.String(),
				Name:            evaluator.Model.Name,
				MaxOutputTokens: derefInt(evaluator.Model.MaxOutputTokens),
			},
			Bounds: EvaluatorBoundsConfig{
				MaxModelTurns:  evaluator.Bounds.MaxModelTurns,
				MaxToolCalls:   evaluator.Bounds.MaxToolCalls,
				TimeoutSeconds: evaluator.Bounds.TimeoutSeconds,
			},
			Retry: RetryPolicyConfig{
				MaxAttempts:                evaluator.Retry.MaxAttempts,
				RetryOnModelError:          evaluator.Retry.RetryOnModelError,
				RetryOnToolFailure:         evaluator.Retry.RetryOnToolFailure,
				RetryOnFinalizationFailure: evaluator.Retry.RetryOnFinalizationFailure,
				RetryOnInvalidPrediction:   evaluator.Retry.RetryOnInvalidPrediction,
			},
		},
		Scoring: ScoringConfig{
			ObjectivePath: objectivePath,
			CurrentEvidence: score.ObjectiveEvidenceRef{
				Name:         "current",
				BundlePath:   string(expectedBundlePath),
				EvidencePath: filepath.Join(string(expectedBundlePath), "evidence.pkl"),
				ReportPath:   filepath.Join(string(expectedBundlePath), "round-report.json"),
			},
			ParentEvidence:     cloneEvidenceRef(request.ParentRef),
			ParentEvidencePath: request.ParentEvidencePath,
		},
		Output: OutputConfig{
			BundleCollectionPath: bundleCollection,
			BundleWriterRoot:     bundleWriterRoot,
			ExpectedBundlePath:   expectedBundlePath,
			ReportFormats:        reportFormats,
			RenderHumanReport:    renderHumanReport,
			ResolvedPolicyPaths: ResolvedPolicyPaths{
				Incumbent:  filepath.ToSlash(incumbent.policyPath),
				Challenger: filepath.ToSlash(challenger.policyPath),
			},
		},
		Report: ReportConfig{
			Formats: reportFormats,
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

func normalizeEvaluationRequest(request evaluationResolveRequest) evaluationResolveRequest {
	if request.Now == nil {
		request.Now = func() time.Time { return time.Now().UTC() }
	}
	return request
}

func resolveSystem(
	manifestDir string,
	evaluator config.Evaluator,
	system config.System,
	policyArtifact *config.PolicyArtifact,
) (resolvedSystem, error) {
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
	if policyArtifact == nil {
		return out, nil
	}

	policyPath, err := resolveExistingManifestPath(manifestDir, policyArtifact.Path)
	if err != nil {
		return resolvedSystem{}, fmt.Errorf("resolve policy path: %w", err)
	}
	data, err := os.ReadFile(policyPath)
	if err != nil {
		return resolvedSystem{}, fmt.Errorf("read policy source: %w", err)
	}
	policy := domain.NewPythonPolicy(domain.PolicyID(policyArtifact.Id), string(data), selectionPolicyV1DefaultSymbol)
	out.spec.Policy = &policy
	out.policyPath = policyPath
	return out, nil
}

func resolveBundlePaths(manifestDir string, override string) (domain.HostPath, domain.HostPath, error) {
	if strings.TrimSpace(override) != "" {
		writerRoot, err := resolveOverridePath(override)
		if err != nil {
			return "", "", err
		}
		return domain.HostPath(writerRoot), domain.HostPath(writerRoot), nil
	}

	writerRoot := filepath.Join(manifestDir, "artifacts")
	return domain.HostPath(writerRoot), domain.HostPath(writerRoot), nil
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
		return "round"
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

func stringifyReportFormats(formats []config.ReportFormat) []string {
	out := make([]string, 0, len(formats))
	for _, format := range formats {
		out = append(out, format.String())
	}
	return out
}

func containsReportFormat(formats []string, target string) bool {
	for _, format := range formats {
		if format == target {
			return true
		}
	}
	return false
}

func cloneEvidenceRef(ref *score.ObjectiveEvidenceRef) *score.ObjectiveEvidenceRef {
	if ref == nil {
		return nil
	}
	copyRef := *ref
	return &copyRef
}
