package optimizer

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

var ErrUnsupportedMode = errors.New("optimizer: only optimization mode is supported")

// Resolve loads one Pkl optimization manifest into the canonical optimizer plan.
func Resolve(ctx context.Context, request ResolveRequest) (Plan, error) {
	request = normalizeRequest(request)

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
	if cfg.Mode != config.ModeOptimization {
		return Plan{}, fmt.Errorf("%w: %s", ErrUnsupportedMode, cfg.Mode)
	}
	if cfg.Optimization == nil || cfg.Agents.Optimizer == nil {
		return Plan{}, fmt.Errorf("%w: incomplete optimization manifest", ErrUnsupportedMode)
	}

	manifestDir := filepath.Dir(manifestPath)
	parentBundlePath, err := resolveExistingManifestPath(manifestDir, cfg.Optimization.ParentRound.Bundle.Path)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve parent bundle path: %w", err)
	}
	if strings.TrimSpace(request.ParentBundlePathOverride) != "" {
		parentBundlePath, err = filepath.Abs(request.ParentBundlePathOverride)
		if err != nil {
			return Plan{}, fmt.Errorf("resolve parent bundle override: %w", err)
		}
	}
	inputPolicyPath, err := resolveExistingManifestPath(manifestDir, cfg.Optimization.Target.Input.Path)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve input policy path: %w", err)
	}
	bundleCollection, bundleWriterRoot, err := resolveBundlePaths(manifestDir, request.BundleRootOverride)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve bundle root: %w", err)
	}

	now := request.Now().UTC()
	bundleID := request.BundleID
	if bundleID == "" {
		bundleID = defaultBundleID(cfg.Name, now)
	}

	parentBundleID := filepath.Base(parentBundlePath)
	systemPrompt := ""
	if cfg.Agents.Optimizer.SystemPrompt != nil {
		systemPrompt = strings.TrimSpace(*cfg.Agents.Optimizer.SystemPrompt)
	}

	return Plan{
		ManifestPath:       manifestPath,
		ExperimentName:     cfg.Name,
		CreatedAt:          now,
		BundleID:           bundleID,
		BundleCollection:   bundleCollection,
		BundleWriterRoot:   bundleWriterRoot,
		ExpectedBundlePath: filepath.Join(bundleCollection, bundleID),
		Agent: pureoptimizer.AgentConfig{
			Model: pureoptimizer.ModelConfig{
				Provider:        cfg.Agents.Optimizer.Model.Provider.String(),
				Name:            cfg.Agents.Optimizer.Model.Name,
				MaxOutputTokens: derefInt(cfg.Agents.Optimizer.Model.MaxOutputTokens),
			},
			Bounds: pureoptimizer.Bounds{
				MaxModelTurns:  cfg.Agents.Optimizer.Bounds.MaxModelTurns,
				MaxToolCalls:   cfg.Agents.Optimizer.Bounds.MaxToolCalls,
				TimeoutSeconds: cfg.Agents.Optimizer.Bounds.TimeoutSeconds,
			},
			Tools: pureoptimizer.ToolPolicy{
				Allow: stringifyTools(cfg.Agents.Optimizer.Tools.Allow),
				Deny:  stringifyTools(cfg.Agents.Optimizer.Tools.Deny),
			},
			SystemPrompt: systemPrompt,
		},
		Target: pureoptimizer.Target{
			InputArtifactID:  artifactID(cfg.Optimization.Target.Input.Id),
			OutputArtifactID: artifactID(cfg.Optimization.Target.Output.Id),
			OutputName:       cfg.Optimization.Target.Output.ArtifactName,
			InterfaceID:      cfg.Optimization.Target.Output.Implements.Id,
		},
		ParentBundle: pureoptimizer.ParentRoundRef{
			ArtifactID: artifactID(cfg.Optimization.ParentRound.Bundle.Id),
			BundleID:   parentBundleID,
			BundlePath: hostPath(parentBundlePath),
		},
		InputPolicy: InputPolicyPlan{
			ArtifactID:  cfg.Optimization.Target.Input.Id,
			Path:        inputPolicyPath,
			InterfaceID: cfg.Optimization.Target.Input.Implements.Id,
		},
		IncludedEvidence: stringifyEvidenceKinds(cfg.Optimization.Evidence.Include),
		DeniedEvidence:   stringifyDeniedEvidenceKinds(cfg.Optimization.Evidence.Deny),
	}, nil
}

func normalizeRequest(request ResolveRequest) ResolveRequest {
	if request.Now == nil {
		request.Now = func() time.Time { return time.Now().UTC() }
	}
	return request
}

func resolveExistingManifestPath(manifestDir string, relPath string) (string, error) {
	joined := filepath.Join(manifestDir, relPath)
	path, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	return path, nil
}

func resolveBundlePaths(manifestDir string, override string) (string, string, error) {
	if strings.TrimSpace(override) != "" {
		path, err := filepath.Abs(override)
		if err != nil {
			return "", "", err
		}
		if filepath.Base(path) == "runs" {
			return path, filepath.Dir(path), nil
		}
		return filepath.Join(path, "runs"), path, nil
	}

	writerRoot := filepath.Join(manifestDir, "artifacts")
	return filepath.Join(writerRoot, "runs"), writerRoot, nil
}

func defaultBundleID(name string, now time.Time) string {
	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
	slug = strings.ReplaceAll(slug, "_", "-")
	return fmt.Sprintf("%s-%s", slug, now.Format("20060102-150405"))
}

func stringifyTools(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func stringifyEvidenceKinds(values []config.OptimizerEvidenceKind) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.String())
	}
	return out
}

func stringifyDeniedEvidenceKinds(values []config.OptimizerDeniedEvidenceKind) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.String())
	}
	return out
}

func derefInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func artifactID(value string) domain.ArtifactID {
	return domain.ArtifactID(value)
}

func hostPath(value string) domain.HostPath {
	return domain.HostPath(value)
}
