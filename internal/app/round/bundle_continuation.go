package round

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/report"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

func roundBundleArtifacts(plan Plan) ([]bundlefs.BundleArtifact, map[domain.Role]string) {
	artifacts := make([]bundlefs.BundleArtifact, 0, 2)
	paths := make(map[domain.Role]string, 2)

	if plan.Policies.Incumbent.Policy != nil {
		path := "policies/incumbent_policy.py"
		artifacts = append(artifacts, bundlefs.BundleArtifact{
			Kind:      "policy",
			Path:      path,
			MediaType: "text/x-python",
			Content:   []byte(plan.Policies.Incumbent.Policy.Source),
		})
		paths[domain.RoleIncumbent] = path
	}
	if plan.Policies.Challenger.Policy != nil {
		name := plan.ChallengerMaterialization.ArtifactName
		if strings.TrimSpace(name) == "" {
			name = "challenger_policy.py"
		}
		path := filepath.ToSlash(filepath.Join("policies", filepath.Base(name)))
		artifacts = append(artifacts, bundlefs.BundleArtifact{
			Kind:      "policy",
			Path:      path,
			MediaType: "text/x-python",
			Content:   []byte(plan.Policies.Challenger.Policy.Source),
		})
		paths[domain.RoleChallenger] = path
	}

	return artifacts, paths
}

func buildContinuation(
	plan Plan,
	roundReport report.RoundReport,
	objective *score.ObjectiveResult,
	policyPaths map[domain.Role]string,
) *pureround.Continuation {
	survivorRole := continuationSurvivorRole(plan)
	survivor := plan.Policies.ByRole(survivorRole)

	continuation := &pureround.Continuation{
		SchemaVersion: pureround.ContinuationSchemaVersion,
		BundleID:      plan.Bundle.ID,
		Game: pureround.ContinuationGame{
			ID:   plan.Game.ID,
			Kind: plan.Game.Kind,
		},
		Round: pureround.ContinuationRound{
			ID: plan.Round.ID,
		},
		CandidateInterface: pureround.ContinuationInterface{
			ID: plan.CandidateInterfaceID,
		},
		SurvivingCandidate: pureround.ContinuationCandidate{
			Role:         survivorRole,
			System:       survivor,
			ArtifactPath: policyPaths[survivorRole],
		},
		DefaultContinuation: pureround.ContinuationDefaults{
			IncumbentFrom: "surviving_candidate",
			MatchesFrom:   "continuation",
			ObjectiveFrom: "continuation",
			EvaluatorFrom: "continuation",
		},
		Dataset: pureround.ContinuationDataset{
			Kind:     plan.Dataset.Kind,
			Name:     plan.Dataset.Name,
			Config:   plan.Dataset.Config,
			Split:    plan.Dataset.Split,
			MaxItems: plan.Dataset.MaxItems,
		},
		Matches: plan.Matches,
		Evaluator: pureround.ContinuationEvaluator{
			Model: pureround.ContinuationModel{
				Provider:        plan.Evaluator.Model.Provider,
				Name:            plan.Evaluator.Model.Name,
				MaxOutputTokens: plan.Evaluator.Model.MaxOutputTokens,
			},
			Bounds: pureround.ContinuationBounds{
				MaxModelTurns:  plan.Evaluator.Bounds.MaxModelTurns,
				MaxToolCalls:   plan.Evaluator.Bounds.MaxToolCalls,
				TimeoutSeconds: plan.Evaluator.Bounds.TimeoutSeconds,
			},
			Retry: pureround.ContinuationRetry{
				MaxAttempts:                plan.Evaluator.Retry.MaxAttempts,
				RetryOnModelError:          plan.Evaluator.Retry.RetryOnModelError,
				RetryOnToolFailure:         plan.Evaluator.Retry.RetryOnToolFailure,
				RetryOnFinalizationFailure: plan.Evaluator.Retry.RetryOnFinalizationFailure,
				RetryOnInvalidPrediction:   plan.Evaluator.Retry.RetryOnInvalidPrediction,
			},
			AllowedTools: append([]string(nil), plan.Evaluator.ToolPolicy.EffectiveAllowed...),
			DeniedTools:  append([]string(nil), plan.Evaluator.ToolPolicy.Denied...),
			SystemPrompt: plan.Evaluator.ToolPolicy.SystemPrompt,
			PolicySHA256: plan.Evaluator.ToolPolicy.PolicySHA256,
		},
		Scoring: pureround.ContinuationScoring{
			ObjectivePath: plan.Scoring.ObjectivePath,
			ReportFormats: append([]string(nil), plan.Report.Formats...),
		},
		Evidence: pureround.ContinuationEvidence{
			RoundEvidencePath: "evidence.pkl",
			ReportPath:        "round-report.json",
			DecisionPath:      "decision.json",
		},
	}
	if objective != nil {
		continuation.Evidence.ObjectivePath = "objective.json"
	}
	_ = roundReport
	return continuation
}

func buildContinuationPKLInput(plan Plan) (*bundlefs.ContinuationPKLInput, error) {
	schemaPath, helpersPath, err := resolveGameSchemaPaths(plan.ManifestPath, plan.Game.ID)
	if err != nil {
		return nil, err
	}
	return &bundlefs.ContinuationPKLInput{
		Name:        plan.Round.ID + "-continuation",
		SchemaPath:  schemaPath,
		HelpersPath: helpersPath,
	}, nil
}

func resolveGameSchemaPaths(manifestPath string, gameID string) (string, string, error) {
	if strings.TrimSpace(manifestPath) == "" {
		return "", "", fmt.Errorf("resolve continuation schema paths: manifest path is required")
	}
	if strings.TrimSpace(gameID) == "" {
		return "", "", fmt.Errorf("resolve continuation schema paths: game id is required")
	}
	startDir := filepath.Dir(manifestPath)
	schemaRelPath := filepath.Join("configs", "schema", "games", gameID+".pkl")
	helpersRelPath := filepath.Join("configs", "schema", "games", gameID+"-helpers.pkl")
	for dir := startDir; ; dir = filepath.Dir(dir) {
		schemaPath := filepath.Join(dir, schemaRelPath)
		helpersPath := filepath.Join(dir, helpersRelPath)
		if _, err := os.Stat(schemaPath); err == nil {
			if _, helperErr := os.Stat(helpersPath); helperErr == nil {
				return filepath.Clean(schemaPath), filepath.Clean(helpersPath), nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", "", fmt.Errorf("resolve continuation schema paths: schema files not found for game %q", gameID)
}

func continuationSurvivorRole(plan Plan) domain.Role {
	if plan.Policies.Challenger.Policy != nil {
		return domain.RoleChallenger
	}
	if plan.Policies.Incumbent.Policy != nil {
		return domain.RoleIncumbent
	}
	return domain.RoleChallenger
}
