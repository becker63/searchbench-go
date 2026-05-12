package round

import (
	"context"
	"strings"

	bundlefs "github.com/becker63/searchbench-go/internal/adapters/bundle/fs"
	"github.com/becker63/searchbench-go/internal/games/codelocalization"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	"github.com/becker63/searchbench-go/internal/pure/game"
	"github.com/becker63/searchbench-go/internal/pure/report"
	pureround "github.com/becker63/searchbench-go/internal/pure/round"
	"github.com/becker63/searchbench-go/internal/pure/score"
)

// Run completes one round and, when configured, proposes a next challenger.
//
// The flow is the named pipeline: ResolveGame -> ResolveRound ->
// MaterializeChallenger -> EvaluateMatches -> BuildEvidence ->
// EvaluateObjective -> Decide -> WriteBundle. Each step delegates to the
// corresponding phase function, and only WriteBundle is allowed to write the
// COMPLETE marker for the round bundle.
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
	resolved, err = MaterializeChallenger(ctx, resolved, input)
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

	objective, err := EvaluateObjective(ctx, resolved, evidence, matches, input)
	if err != nil {
		return record, err
	}

	decision := Decide(resolvedGame, resolved, evidence, objective, matches)

	bundleRef, err := WriteBundle(ctx, resolved, matches, evidence, objective, input)
	if err != nil {
		return record, err
	}

	roundResult := buildResult(matches, bundleRef, evidence, objective)
	record.Round = pureround.Record{
		GameID:          resolvedGame.ID,
		RoundID:         evidence.RoundID,
		BundlePath:      bundleRef.Path,
		Evidence:        evidence,
		ObjectiveResult: objective,
		Decision:        decision,
	}
	record.RoundBundle = string(bundleRef.Path)
	record.RoundResult = &roundResult
	return record, nil
}

// ResolveGame resolves the domain contract for the round.
func ResolveGame(_ context.Context, _ Input) (game.Contract, error) {
	return codelocalization.Contract(), nil
}

// ResolveRound loads and normalizes the caller's round manifest.
func ResolveRound(ctx context.Context, resolvedGame game.Contract, input Input) (Resolved, error) {
	roundPlan, err := resolveEvaluation(ctx, evaluationResolveRequest{
		ManifestPath:                input.EvaluationManifestPath,
		BundleRootOverride:          roundBundleRoot(input.BundleRootOverride),
		BundleID:                    input.RoundID,
		Now:                         input.Now,
		DatasetMaterializeCacheDir:  domain.HostPath(strings.TrimSpace(input.DatasetMaterializeCacheDir)),
		DatasetMaterializeRemoteURL: strings.TrimSpace(input.DatasetMaterializeRemoteURL),
	})
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{Game: resolvedGame, Round: roundPlan}, nil
}

// MaterializeChallenger resolves any optimizer-backed challenger generation
// before evaluation begins. Checked-in challengers leave the resolved plan
// unchanged.
func MaterializeChallenger(ctx context.Context, resolved Resolved, input Input) (Resolved, error) {
	updated, err := materializeChallenger(ctx, resolved, input)
	if err != nil {
		return Resolved{}, &Error{Phase: PhaseMaterializeChallengerFailed, Err: err}
	}
	return updated, nil
}

// EvaluateMatches runs incumbent and challenger executions for the round
// matches. It must not write any bundle artifacts; persistence is reserved for
// WriteBundle.
func EvaluateMatches(ctx context.Context, resolved Resolved, input Input) (MatchRecords, error) {
	plan := resolved.Round
	runCtx, cancel := withEvaluatorTimeout(ctx, plan)
	defer cancel()

	roundReport, executions, matchExec, err := runComparison(runCtx, plan, evaluationRequestFromInput(input))
	if err != nil {
		return MatchRecords{}, &Error{Phase: PhaseComparisonFailed, Err: err}
	}
	return MatchRecords{
		Plan:                plan,
		RoundReport:         roundReport,
		EvaluatorExecutions: executions,
		MatchExecutions:     matchExec,
	}, nil
}

// BuildEvidence projects the durable round evidence from the match records.
// It must not write any bundle artifacts.
func BuildEvidence(_ game.Contract, _ Resolved, matches MatchRecords) (score.RoundEvidenceDocument, error) {
	evidence, err := projectRoundEvidence(matches.Plan, matches.RoundReport, matches.MatchExecutions)
	if err != nil {
		return score.RoundEvidenceDocument{}, &Error{Phase: PhaseRoundEvidenceFailed, Err: err}
	}
	return evidence, nil
}

// EvaluateObjective scores the projected evidence with the manifest's
// objective. It must not write any bundle artifacts.
func EvaluateObjective(
	ctx context.Context,
	_ Resolved,
	evidence score.RoundEvidenceDocument,
	matches MatchRecords,
	input Input,
) (*score.ObjectiveResult, error) {
	return evaluateObjectiveForPlan(ctx, matches.Plan, evidence, evaluationRequestFromInput(input))
}

// Decide captures the explicit round decision from the round report.
func Decide(_ game.Contract, _ Resolved, _ score.RoundEvidenceDocument, _ *score.ObjectiveResult, matches MatchRecords) report.Decision {
	return matches.RoundReport.Decision
}

// WriteBundle is the only phase that persists the durable round bundle and
// writes the COMPLETE marker.
func WriteBundle(
	ctx context.Context,
	_ Resolved,
	matches MatchRecords,
	evidence score.RoundEvidenceDocument,
	objective *score.ObjectiveResult,
	input Input,
) (bundlefs.BundleRef, error) {
	return writeRoundBundle(
		ctx,
		matches.Plan,
		evaluationRequestFromInput(input),
		matches.RoundReport,
		evidence,
		objective,
	)
}

func buildResult(
	matches MatchRecords,
	bundleRef bundlefs.BundleRef,
	evidence score.RoundEvidenceDocument,
	objective *score.ObjectiveResult,
) Result {
	return Result{
		ManifestPath:        matches.Plan.ManifestPath,
		Bundle:              bundleRef,
		ReportID:            matches.RoundReport.ID,
		RoundReport:         matches.RoundReport,
		RoundEvidence:       evidence,
		ObjectiveResult:     objective,
		EvaluatorExecutions: matches.EvaluatorExecutions,
		MatchExecutions:     matches.MatchExecutions,
	}
}

func evaluationRequestFromInput(input Input) evaluationRequest {
	return evaluationRequest{
		Resolve: evaluationResolveRequest{
			ManifestPath:                input.EvaluationManifestPath,
			BundleRootOverride:          roundBundleRoot(input.BundleRootOverride),
			BundleID:                    input.RoundID,
			Now:                         input.Now,
			DatasetMaterializeCacheDir:  domain.HostPath(strings.TrimSpace(input.DatasetMaterializeCacheDir)),
			DatasetMaterializeRemoteURL: strings.TrimSpace(input.DatasetMaterializeRemoteURL),
		},
		DisableRenderReport:   input.DisableRenderReport,
		EvaluatorModelFactory: input.EvaluatorModelFactory,
		EvaluatorToolFactory:  input.EvaluatorToolFactory,
	}
}
