package round

import (
	"context"
	"fmt"

	"github.com/becker63/searchbench-go/internal/app/evaluation"
	appOptimizer "github.com/becker63/searchbench-go/internal/app/optimizer"
	"github.com/becker63/searchbench-go/internal/games/codelocalization"
	"github.com/becker63/searchbench-go/internal/pure/game"
	"github.com/becker63/searchbench-go/internal/pure/report"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Run completes one round and, when configured, proposes a next challenger.
func Run(ctx context.Context, input Input) (Record, error) {
	input = normalizeInput(input)
	record := Record{}

	resolvedGame, err := ResolveGame(ctx, input)
	if err != nil {
		return record, err
	}
	record.Game = resolvedGame

	resolved, err := ResolveRound(ctx, resolvedGame, input)
	if err != nil {
		return record, err
	}

	matches, err := EvaluateMatches(ctx, resolved, input)
	if err != nil {
		return record, err
	}

	evidence, err := BuildEvidence(resolvedGame, resolved, matches)
	if err != nil {
		return record, err
	}

	objective, err := EvaluateObjective(ctx, resolved, evidence, matches)
	if err != nil {
		return record, err
	}

	decision := Decide(resolvedGame, resolved, evidence, objective, matches)

	next, err := ProposeNextChallenger(ctx, resolved, evidence, objective, decision, matches, input)
	if err != nil {
		return record, err
	}

	record.Round = WriteBundle(resolved, evidence, objective, decision, next, matches)
	record.RoundBundle = string(matches.Evaluation.Bundle.Path)
	record.RoundResult = &matches.Evaluation
	if next != nil {
		record.NextChallengerResult = next
		record.OptimizerBundle = next.BundlePath
	}
	return record, nil
}

// ResolveGame resolves the domain contract for the round.
func ResolveGame(_ context.Context, _ Input) (game.Contract, error) {
	return codelocalization.Contract(), nil
}

// ResolveRound loads and normalizes the caller's round manifest.
func ResolveRound(ctx context.Context, resolvedGame game.Contract, input Input) (Resolved, error) {
	roundPlan, err := evaluation.Resolve(ctx, evaluation.ResolveRequest{
		ManifestPath:       input.EvaluationManifestPath,
		BundleRootOverride: roundBundleRoot(input.BundleRootOverride),
		BundleID:           input.RoundID,
		Now:                input.Now,
	})
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{Game: resolvedGame, Round: roundPlan}, nil
}

// EvaluateMatches runs incumbent and challenger executions for the round matches.
func EvaluateMatches(ctx context.Context, resolved Resolved, input Input) (MatchRecords, error) {
	roundResult, err := evaluation.RunResolved(ctx, resolved.Round, evaluation.Request{
		Resolve: evaluation.ResolveRequest{
			ManifestPath:       input.EvaluationManifestPath,
			BundleRootOverride: roundBundleRoot(input.BundleRootOverride),
			BundleID:           input.RoundID,
			Now:                input.Now,
		},
		EvaluatorModelFactory: input.EvaluatorModelFactory,
		EvaluatorToolFactory:  input.EvaluatorToolFactory,
	})
	if err != nil {
		return MatchRecords{}, err
	}
	return MatchRecords{Evaluation: roundResult}, nil
}

// BuildEvidence returns the durable round evidence derived from match records.
func BuildEvidence(_ game.Contract, _ Resolved, matches MatchRecords) (score.RoundEvidenceDocument, error) {
	return matches.Evaluation.RoundEvidence, nil
}

// EvaluateObjective returns the objective result already evaluated for the round.
func EvaluateObjective(_ context.Context, _ Resolved, _ score.RoundEvidenceDocument, matches MatchRecords) (*score.ObjectiveResult, error) {
	return matches.Evaluation.ObjectiveResult, nil
}

// Decide captures the explicit round decision from the round report.
func Decide(_ game.Contract, _ Resolved, _ score.RoundEvidenceDocument, _ *score.ObjectiveResult, matches MatchRecords) report.Decision {
	return matches.Evaluation.RoundReport.Decision
}

// ProposeNextChallenger asks the optimizer for a possible future challenger.
func ProposeNextChallenger(
	ctx context.Context,
	_ Resolved,
	_ score.RoundEvidenceDocument,
	_ *score.ObjectiveResult,
	_ report.Decision,
	matches MatchRecords,
	input Input,
) (*appOptimizer.Record, error) {
	optimizerModelFactory := input.OptimizerModelFactory
	if optimizerModelFactory == nil {
		return nil, fmt.Errorf("round: optimizer model factory is required")
	}
	optimizerModel, err := optimizerModelFactory()
	if err != nil {
		return nil, err
	}

	optimizerPlan, err := appOptimizer.Resolve(ctx, appOptimizer.ResolveRequest{
		ManifestPath:             input.OptimizationManifestPath,
		BundleRootOverride:       optimizerBundleRoot(input.BundleRootOverride),
		ParentBundlePathOverride: string(matches.Evaluation.Bundle.Path),
		BundleID:                 input.OptimizerBundleID,
		Now:                      input.Now,
	})
	if err != nil {
		return nil, err
	}

	nextChallengerRecord, err := appOptimizer.RunResolved(ctx, optimizerPlan, appOptimizer.Request{
		Resolve: appOptimizer.ResolveRequest{
			ManifestPath:             input.OptimizationManifestPath,
			BundleRootOverride:       optimizerBundleRoot(input.BundleRootOverride),
			ParentBundlePathOverride: string(matches.Evaluation.Bundle.Path),
			BundleID:                 input.OptimizerBundleID,
			Now:                      input.Now,
		},
		Model: optimizerModel,
	})
	if err != nil {
		return &nextChallengerRecord, err
	}

	return &nextChallengerRecord, nil
}

// WriteBundle records the already-written durable round bundle in the round record.
func WriteBundle(
	resolved Resolved,
	evidence score.RoundEvidenceDocument,
	objective *score.ObjectiveResult,
	decision report.Decision,
	next *appOptimizer.Record,
	matches MatchRecords,
) pureround.Record {
	return pureround.Record{
		GameID:          resolved.Game.ID,
		RoundID:         evidence.RoundID,
		BundlePath:      matches.Evaluation.Bundle.Path,
		Evidence:        evidence,
		ObjectiveResult: objective,
		Decision:        decision,
		NextChallenger:  next != nil && next.Optimizer.Proposal != nil,
	}
}
