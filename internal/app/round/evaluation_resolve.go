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
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/ports/dataset"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// defaultMatchSource is the deterministic local match source used when the
// caller does not inject a real dataset adapter. It is wired through the
// MatchSource port so a real adapter can be substituted without touching
// resolveEvaluation.
var defaultMatchSource dataset.MatchSource = evaluatorfake.NewMatchSource()

// selectionPolicyV1DefaultSymbol is the runtime callable used for policy
// artifacts that implement iterative_context.selection_policy.v1.
const selectionPolicyV1DefaultSymbol = "score"

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
	if cfg.Round == nil {
		return Plan{}, fmt.Errorf("resolve round surface: round block is required")
	}
	return resolveRoundManifest(ctx, cfg, manifestPath, request)
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
