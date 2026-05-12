package round

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	config "github.com/becker63/searchbench-go/internal/adapters/config/pkl"
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	evaluatorpolicy "github.com/becker63/searchbench-go/internal/agents/evaluator/policy"
	optimizepolicy "github.com/becker63/searchbench-go/internal/agents/optimizer/policy"
	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/ports/dataset"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func resolveRoundManifest(ctx context.Context, cfg config.RoundSpec, manifestPath string, request evaluationResolveRequest) (Plan, error) {
	round := cfg.Round
	if round == nil {
		return Plan{}, fmt.Errorf("resolve round surface: round block is required")
	}

	manifestDir := filepath.Dir(manifestPath)
	now := request.Now().UTC()
	bundleID := request.BundleID
	if bundleID == "" {
		bundleID = round.Id
	}
	if bundleID == "" {
		bundleID = defaultBundleID(cfg.Name, now)
	}
	reportID := request.ReportID
	if reportID.Empty() {
		reportID = defaultReportID(bundleID)
	}

	_, bundleWriterRoot, err := resolveBundlePaths(manifestDir, request.BundleRootOverride)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve bundle root: %w", err)
	}
	gameID := cfg.Game.Id
	if strings.TrimSpace(gameID) == "" {
		gameID = "code-localization"
	}
	bundleCollection := domain.HostPath(filepath.Join(string(bundleWriterRoot), "games", gameID, "rounds"))
	expectedBundlePath := domain.HostPath(filepath.Join(string(bundleCollection), bundleID))

	var (
		parentContinuation *pureround.Continuation
		parentBundlePath   string
		parentEvidenceRef  *score.ObjectiveEvidenceRef
		parentEvidencePath string
	)
	if round.Continues != nil && strings.TrimSpace(*round.Continues) != "" {
		parentBundlePath, err = resolveContinuationBundlePath(manifestPath, manifestDir, *round.Continues)
		if err != nil {
			return Plan{}, fmt.Errorf("resolve continuation bundle path: %w", err)
		}
		continuation, err := bundlefs.LoadContinuation(domain.HostPath(parentBundlePath))
		if err != nil {
			return Plan{}, fmt.Errorf("load continuation bundle: %w", err)
		}
		parentContinuation = &continuation
		parentEvidenceRef = &score.ObjectiveEvidenceRef{
			Name:         "parent",
			BundlePath:   parentBundlePath,
			EvidencePath: filepath.Join(parentBundlePath, "evidence.pkl"),
			ReportPath:   filepath.Join(parentBundlePath, "round-report.json"),
		}
		parentEvidencePath = filepath.Join(parentBundlePath, "evidence.pkl")
	}

	matches, datasetConfig, err := resolveRoundMatches(ctx, manifestDir, cfg, round, parentContinuation, request)
	if err != nil {
		return Plan{}, err
	}
	evaluatorConfig, evaluatorModelSpec, evaluatorBounds, evaluatorMaxOutputTokens, err := resolveRoundEvaluator(round, parentContinuation)
	if err != nil {
		return Plan{}, err
	}
	scoringConfig, err := resolveRoundScoring(manifestDir, round, parentContinuation, expectedBundlePath, parentEvidenceRef, parentEvidencePath)
	if err != nil {
		return Plan{}, err
	}

	incumbent, incumbentPolicyPath, err := resolveRoundIncumbent(
		manifestDir,
		round,
		parentContinuation,
		parentBundlePath,
		evaluatorModelSpec,
		evaluatorBounds,
		evaluatorMaxOutputTokens,
	)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve incumbent policy: %w", err)
	}
	challenger, challengerPolicyPath, materialization, optimizerConfig, candidateInterfaceID, err := resolveRoundChallenger(
		manifestDir,
		round,
		parentContinuation,
		parentBundlePath,
		evaluatorModelSpec,
		evaluatorBounds,
		evaluatorMaxOutputTokens,
	)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve challenger policy: %w", err)
	}

	systems := domain.NewPair(incumbent, challenger)
	comparePlan := compare.NewPlan(systems, matches)
	if err := comparePlan.Validate(); err != nil {
		return Plan{}, fmt.Errorf("validate compare plan: %w", err)
	}

	reportFormats := stringifyReportFormats(round.Report.Formats)
	renderHumanReport := containsReportFormat(reportFormats, config.ReportFormatText.String())
	if parentContinuation != nil && len(reportFormats) == 0 {
		reportFormats = append(reportFormats, parentContinuation.Scoring.ReportFormats...)
		renderHumanReport = containsReportFormat(reportFormats, config.ReportFormatText.String())
	}

	roundName := cfg.Name
	if strings.TrimSpace(roundName) == "" {
		roundName = round.Id
	}

	plan := Plan{
		ManifestPath:         manifestPath,
		RoundName:            roundName,
		Mode:                 "round",
		Game:                 GameConfig{ID: cfg.Game.Id, Kind: cfg.Game.Kind},
		Round:                RoundConfig{ID: bundleID},
		Dataset:              datasetConfig,
		Policies:             systems,
		Matches:              matches,
		Parallelism:          compare.DefaultParallelism(),
		CandidateInterfaceID: candidateInterfaceID,
		Evaluator:            evaluatorConfig,
		Optimizer:            optimizerConfig,
		Scoring:              scoringConfig,
		Output: OutputConfig{
			BundleCollectionPath: bundleCollection,
			BundleWriterRoot:     bundleWriterRoot,
			ExpectedBundlePath:   expectedBundlePath,
			ReportFormats:        reportFormats,
			RenderHumanReport:    renderHumanReport,
			ResolvedPolicyPaths: ResolvedPolicyPaths{
				Incumbent:  filepath.ToSlash(incumbentPolicyPath),
				Challenger: filepath.ToSlash(challengerPolicyPath),
			},
		},
		Report: ReportConfig{
			Formats: reportFormats,
		},
		Bundle: BundleConfig{
			ID: bundleID,
		},
		Lineage: LineageConfig{
			Continues: parentBundlePath,
		},
		ChallengerMaterialization: materialization,
		ReportID:                  reportID,
		CreatedAt:                 now,
	}
	return plan, nil
}

func resolveContinuationBundlePath(manifestPath string, manifestDir string, continues string) (string, error) {
	if strings.TrimSpace(continues) == "." {
		if parentBundlePath, ok, err := parentBundlePathFromAmends(manifestPath); err != nil {
			return "", err
		} else if ok {
			return parentBundlePath, nil
		}
	}
	return resolveExistingManifestPath(manifestDir, continues)
}

func parentBundlePathFromAmends(manifestPath string) (string, bool, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", false, err
	}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if !strings.HasPrefix(line, "amends ") {
			return "", false, nil
		}
		target, ok := pklQuotedPath(line[len("amends "):])
		if !ok {
			return "", false, nil
		}
		resolved, err := resolveExistingManifestPath(filepath.Dir(manifestPath), target)
		if err != nil {
			return "", false, err
		}
		if filepath.Base(resolved) != bundlefsContinuationPKLFileName() {
			return "", false, nil
		}
		return filepath.Dir(resolved), true, nil
	}
	return "", false, nil
}

func pklQuotedPath(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", false
	}
	inner := value[1 : len(value)-1]
	inner = strings.ReplaceAll(inner, `\"`, `"`)
	inner = strings.ReplaceAll(inner, `\\`, `\`)
	return inner, true
}

func bundlefsContinuationPKLFileName() string {
	return "continuation.pkl"
}

func resolveRoundMatches(
	ctx context.Context,
	manifestDir string,
	cfg config.RoundSpec,
	round *config.RoundManifest,
	parent *pureround.Continuation,
	request evaluationResolveRequest,
) (domain.NonEmpty[domain.MatchSpec], DatasetConfig, error) {
	if round != nil && round.Matches != nil {
		// Materialize matches via dataset.MatchSource (JetBrains LCA JSONL under
		// the manifest directory, or the local fake for other selections).
		matches, err := defaultMatchSource.Matches(ctx, dataset.Request{
			ManifestDir:          manifestDir,
			Kind:                 round.Matches.Kind,
			Name:                 round.Matches.Name,
			Config:               round.Matches.Config,
			Split:                round.Matches.Split,
			MaxItems:             round.Matches.MaxItems,
			MaterializeCacheDir:  request.DatasetMaterializeCacheDir,
			MaterializeRemoteURL: request.DatasetMaterializeRemoteURL,
		})
		if err != nil {
			return domain.NonEmpty[domain.MatchSpec]{}, DatasetConfig{}, fmt.Errorf("resolve matches: %w", err)
		}
		return matches, DatasetConfig{
			Kind:     round.Matches.Kind,
			Name:     round.Matches.Name,
			Config:   round.Matches.Config,
			Split:    round.Matches.Split,
			MaxItems: round.Matches.MaxItems,
		}, nil
	}
	if parent != nil {
		return parent.Matches, DatasetConfig{
			Kind:     parent.Dataset.Kind,
			Name:     parent.Dataset.Name,
			Config:   parent.Dataset.Config,
			Split:    parent.Dataset.Split,
			MaxItems: parent.Dataset.MaxItems,
		}, nil
	}
	return domain.NonEmpty[domain.MatchSpec]{}, DatasetConfig{}, fmt.Errorf("resolve matches: missing from-scratch match selection")
}

func resolveRoundEvaluator(round *config.RoundManifest, parent *pureround.Continuation) (EvaluatorConfig, domain.ModelSpec, EvaluatorBoundsConfig, int, error) {
	if round != nil && round.Evaluator != nil {
		effectiveTools, deniedTools, systemPrompt, policyHash, err := evaluatorpolicy.ResolveEvaluatorRunPolicy(
			round.Evaluator.Tools,
			round.Evaluator.SystemPrompt,
			evaluatorpolicy.EvaluatorToolRegistry{
				Available:      evaluatorfake.LocalEvaluatorToolNames(),
				DefaultAllowed: evaluatorfake.LocalEvaluatorDefaultAllowedToolNames(),
			},
		)
		if err != nil {
			return EvaluatorConfig{}, domain.ModelSpec{}, EvaluatorBoundsConfig{}, 0, fmt.Errorf("evaluator tool policy: %w", err)
		}
		configView := EvaluatorConfig{
			Model: EvaluatorModelConfig{
				Provider:        round.Evaluator.Model.Provider.String(),
				Name:            round.Evaluator.Model.Name,
				MaxOutputTokens: derefInt(round.Evaluator.Model.MaxOutputTokens),
			},
			Bounds: EvaluatorBoundsConfig{
				MaxModelTurns:  round.Evaluator.Bounds.MaxModelTurns,
				MaxToolCalls:   round.Evaluator.Bounds.MaxToolCalls,
				TimeoutSeconds: round.Evaluator.Bounds.TimeoutSeconds,
			},
			Retry: RetryPolicyConfig{
				MaxAttempts:                round.Evaluator.Retry.MaxAttempts,
				RetryOnModelError:          round.Evaluator.Retry.RetryOnModelError,
				RetryOnToolFailure:         round.Evaluator.Retry.RetryOnToolFailure,
				RetryOnFinalizationFailure: round.Evaluator.Retry.RetryOnFinalizationFailure,
				RetryOnInvalidPrediction:   round.Evaluator.Retry.RetryOnInvalidPrediction,
			},
			ToolPolicy: EvaluatorToolPolicyView{
				EffectiveAllowed: effectiveTools,
				Denied:           deniedTools,
				SystemPrompt:     systemPrompt,
				PolicySHA256:     policyHash,
			},
		}
		return configView, domain.ModelSpec{
			Provider: configView.Model.Provider,
			Name:     configView.Model.Name,
		}, configView.Bounds, configView.Model.MaxOutputTokens, nil
	}
	if parent != nil {
		configView := EvaluatorConfig{
			Model: EvaluatorModelConfig{
				Provider:        parent.Evaluator.Model.Provider,
				Name:            parent.Evaluator.Model.Name,
				MaxOutputTokens: parent.Evaluator.Model.MaxOutputTokens,
			},
			Bounds: EvaluatorBoundsConfig{
				MaxModelTurns:  parent.Evaluator.Bounds.MaxModelTurns,
				MaxToolCalls:   parent.Evaluator.Bounds.MaxToolCalls,
				TimeoutSeconds: parent.Evaluator.Bounds.TimeoutSeconds,
			},
			Retry: RetryPolicyConfig{
				MaxAttempts:                parent.Evaluator.Retry.MaxAttempts,
				RetryOnModelError:          parent.Evaluator.Retry.RetryOnModelError,
				RetryOnToolFailure:         parent.Evaluator.Retry.RetryOnToolFailure,
				RetryOnFinalizationFailure: parent.Evaluator.Retry.RetryOnFinalizationFailure,
				RetryOnInvalidPrediction:   parent.Evaluator.Retry.RetryOnInvalidPrediction,
			},
			ToolPolicy: EvaluatorToolPolicyView{
				EffectiveAllowed: append([]string(nil), parent.Evaluator.AllowedTools...),
				Denied:           append([]string(nil), parent.Evaluator.DeniedTools...),
				SystemPrompt:     parent.Evaluator.SystemPrompt,
				PolicySHA256:     parent.Evaluator.PolicySHA256,
			},
		}
		return configView, domain.ModelSpec{Provider: configView.Model.Provider, Name: configView.Model.Name}, configView.Bounds, configView.Model.MaxOutputTokens, nil
	}
	return EvaluatorConfig{}, domain.ModelSpec{}, EvaluatorBoundsConfig{}, 0, fmt.Errorf("resolve evaluator: missing effective evaluator")
}

func resolveRoundScoring(
	manifestDir string,
	round *config.RoundManifest,
	parent *pureround.Continuation,
	expectedBundlePath domain.HostPath,
	parentEvidenceRef *score.ObjectiveEvidenceRef,
	parentEvidencePath string,
) (ScoringConfig, error) {
	if round != nil && round.Scoring != nil {
		objectivePath, err := resolveExistingManifestPath(manifestDir, round.Scoring.Objective)
		if err != nil {
			return ScoringConfig{}, fmt.Errorf("resolve objective path: %w", err)
		}
		return ScoringConfig{
			ObjectivePath: objectivePath,
			CurrentEvidence: score.ObjectiveEvidenceRef{
				Name:         "current",
				BundlePath:   string(expectedBundlePath),
				EvidencePath: filepath.Join(string(expectedBundlePath), "evidence.pkl"),
				ReportPath:   filepath.Join(string(expectedBundlePath), "round-report.json"),
			},
			ParentEvidence:     cloneEvidenceRef(parentEvidenceRef),
			ParentEvidencePath: parentEvidencePath,
		}, nil
	}
	if parent != nil {
		return ScoringConfig{
			ObjectivePath: parent.Scoring.ObjectivePath,
			CurrentEvidence: score.ObjectiveEvidenceRef{
				Name:         "current",
				BundlePath:   string(expectedBundlePath),
				EvidencePath: filepath.Join(string(expectedBundlePath), "evidence.pkl"),
				ReportPath:   filepath.Join(string(expectedBundlePath), "round-report.json"),
			},
			ParentEvidence:     cloneEvidenceRef(parentEvidenceRef),
			ParentEvidencePath: parentEvidencePath,
		}, nil
	}
	return ScoringConfig{}, fmt.Errorf("resolve scoring: missing effective objective")
}

func resolveRoundIncumbent(
	manifestDir string,
	round *config.RoundManifest,
	parent *pureround.Continuation,
	parentBundlePath string,
	model domain.ModelSpec,
	bounds EvaluatorBoundsConfig,
	modelMaxOutputTokens int,
) (domain.SystemSpec, string, error) {
	if round != nil && round.Incumbent != nil {
		resolved, err := resolveRoundSystem(manifestDir, round.Incumbent.System, round.Incumbent.SelectionPolicy, model, bounds, modelMaxOutputTokens)
		return resolved.spec, resolved.policyPath, err
	}
	if parent != nil {
		return parent.SurvivingCandidate.System, parent.ResolveArtifactPath(domain.HostPath(parentBundlePath)), nil
	}
	return domain.SystemSpec{}, "", fmt.Errorf("missing effective incumbent")
}

func resolveRoundChallenger(
	manifestDir string,
	round *config.RoundManifest,
	parent *pureround.Continuation,
	parentBundlePath string,
	model domain.ModelSpec,
	bounds EvaluatorBoundsConfig,
	modelMaxOutputTokens int,
) (domain.SystemSpec, string, ChallengerMaterializationConfig, *OptimizerConfig, string, error) {
	var baseSystem domain.SystemSpec
	var basePolicyPath string
	if round.Challenger.System != nil {
		resolved, err := resolveRoundSystem(manifestDir, *round.Challenger.System, nil, model, bounds, modelMaxOutputTokens)
		if err != nil {
			return domain.SystemSpec{}, "", ChallengerMaterializationConfig{}, nil, "", err
		}
		baseSystem = resolved.spec
		basePolicyPath = resolved.policyPath
	} else if parent != nil {
		baseSystem = parent.SurvivingCandidate.System
		baseSystem.Policy = nil
		basePolicyPath = parent.ResolveArtifactPath(domain.HostPath(parentBundlePath))
	} else {
		return domain.SystemSpec{}, "", ChallengerMaterializationConfig{}, nil, "", fmt.Errorf("missing challenger system")
	}

	candidateInterfaceID := ""
	if round.Challenger.SelectionPolicy != nil {
		systemCfg := config.System{
			Id:      string(baseSystem.ID),
			Name:    baseSystem.Name,
			Backend: config.Backend(baseSystem.Backend),
			PromptBundle: config.PromptBundle{
				Name:    baseSystem.PromptBundle.Name,
				Version: roundStringPtr(baseSystem.PromptBundle.Version),
			},
			Runtime: config.Runtime{
				MaxSteps:       baseSystem.Runtime.MaxSteps,
				TimeoutSeconds: bounds.TimeoutSeconds,
			},
		}
		if round.Challenger.System != nil {
			systemCfg = *round.Challenger.System
		}
		resolved, err := resolveRoundSystem(manifestDir, systemCfg, round.Challenger.SelectionPolicy, model, bounds, modelMaxOutputTokens)
		if err != nil {
			return domain.SystemSpec{}, "", ChallengerMaterializationConfig{}, nil, "", err
		}
		candidateInterfaceID = round.Challenger.SelectionPolicy.Implements.Id
		return resolved.spec, resolved.policyPath, ChallengerMaterializationConfig{Mode: "checked_in"}, nil, candidateInterfaceID, nil
	}
	if round.Challenger.Generate != nil {
		if parent == nil {
			return domain.SystemSpec{}, "", ChallengerMaterializationConfig{}, nil, "", fmt.Errorf("generated challenger requires a continuation parent bundle")
		}
		artifactName := round.Challenger.Generate.ArtifactName
		candidateInterfaceID = parent.CandidateInterface.ID
		baseSystem.Policy = nil
		return baseSystem, basePolicyPath, ChallengerMaterializationConfig{
				Mode:             "generated",
				ArtifactName:     artifactName,
				IncludedEvidence: stringifyEvidenceKinds(round.Challenger.Generate.Include),
				DeniedEvidence:   stringifyDeniedEvidenceKinds(round.Challenger.Generate.Deny),
			}, &OptimizerConfig{
				Agent: pureoptimizer.AgentConfig{
					Model: pureoptimizer.ModelConfig{
						Provider:        round.Challenger.Generate.Optimizer.Model.Provider.String(),
						Name:            round.Challenger.Generate.Optimizer.Model.Name,
						MaxOutputTokens: derefInt(round.Challenger.Generate.Optimizer.Model.MaxOutputTokens),
					},
					Bounds: pureoptimizer.Bounds{
						MaxModelTurns:  round.Challenger.Generate.Optimizer.Bounds.MaxModelTurns,
						MaxToolCalls:   round.Challenger.Generate.Optimizer.Bounds.MaxToolCalls,
						TimeoutSeconds: round.Challenger.Generate.Optimizer.Bounds.TimeoutSeconds,
					},
					Tools: pureoptimizer.ToolPolicy{
						Allow: stringifyTools(round.Challenger.Generate.Optimizer.Tools.Allow),
						Deny:  stringifyTools(round.Challenger.Generate.Optimizer.Tools.Deny),
					},
					SystemPrompt: derefString(round.Challenger.Generate.Optimizer.SystemPrompt),
				},
			}, candidateInterfaceID, nil
	}
	return domain.SystemSpec{}, basePolicyPath, ChallengerMaterializationConfig{}, nil, "", fmt.Errorf("missing challenger policy patch")
}

func resolveRoundSystem(
	manifestDir string,
	system config.System,
	policyArtifact *config.PolicyArtifact,
	model domain.ModelSpec,
	bounds EvaluatorBoundsConfig,
	modelMaxOutputTokens int,
) (resolvedSystem, error) {
	backendKind, err := mapBackend(system.Backend)
	if err != nil {
		return resolvedSystem{}, err
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
			MaxSteps:        resolvedMaxSteps(bounds.MaxModelTurns, system.Runtime.MaxSteps),
			MaxToolCalls:    bounds.MaxToolCalls,
			MaxOutputTokens: modelMaxOutputTokens,
		},
	}
	if policyArtifact == nil {
		return resolvedSystem{spec: spec}, nil
	}
	policyPath, err := resolveExistingManifestPath(manifestDir, policyArtifact.Path)
	if err != nil {
		return resolvedSystem{}, fmt.Errorf("resolve policy path: %w", err)
	}
	data, err := os.ReadFile(policyPath)
	if err != nil {
		return resolvedSystem{}, fmt.Errorf("read policy source: %w", err)
	}
	policy := domain.NewPythonPolicy(domain.PolicyID(policyArtifact.Id), string(data), optimizepolicy.CanonicalICPolicySymbol)
	spec.Policy = &policy
	return resolvedSystem{spec: spec, policyPath: policyPath}, nil
}

func roundStringPtr(value string) *string {
	return &value
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

func stringifyEvidenceKinds(values []config.NextChallengerEvidenceKind) []string {
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
