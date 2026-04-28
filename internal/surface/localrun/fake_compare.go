package localrun

import (
	"context"
	"fmt"
	"time"

	"github.com/becker63/searchbench-go/internal/codegraph"
	"github.com/becker63/searchbench-go/internal/compare"
	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
	"github.com/becker63/searchbench-go/internal/surface/logging"
)

func runFakeComparison(ctx context.Context, projected projectedRun) (report.CandidateReport, error) {
	runner := compare.Runner{
		Executor:      fakeExecutor{now: projected.createdAt},
		GraphProvider: fakeGraphProvider{},
		Scorer:        fakeScorer{},
		Decider:       fakeDecider{},
		NewRunID: func(role domain.Role, task domain.TaskSpec, system domain.SystemRef) domain.RunID {
			return domain.RunID(fmt.Sprintf("%s-%s-%s", role, task.ID, system.ID))
		},
		NewReportID: func() domain.ReportID {
			return projected.reportID
		},
		Now: func() time.Time {
			return projected.createdAt
		},
		Parallelism: compare.DefaultParallelism(),
		Logger:      logging.NewNop(),
	}
	out, err := runner.Run(ctx, projected.plan)
	if err != nil {
		return report.CandidateReport{}, err
	}
	out.CreatedAt = projected.createdAt
	return out, nil
}

type fakeExecutor struct {
	now time.Time
}

func (f fakeExecutor) Execute(_ context.Context, spec run.Spec) (run.ExecutedRun, error) {
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("local-session-"+spec.ID.String()))

	predictionFile := domain.RepoRelPath("src/baseline_guess.go")
	usage := domain.UsageSummary{
		InputTokens:  150,
		OutputTokens: 30,
		TotalTokens:  180,
		CostUSD:      0.20,
	}
	if spec.System.Backend == domain.BackendIterativeContext {
		predictionFile = spec.Task.Oracle.GoldFiles[0]
		usage = domain.UsageSummary{
			InputTokens:  40,
			OutputTokens: 12,
			TotalTokens:  52,
			CostUSD:      0.03,
		}
	}

	return run.NewExecuted(
		prepared,
		domain.Prediction{
			Files:     []domain.RepoRelPath{predictionFile},
			Reasoning: "deterministic local fake comparison",
		},
		usage,
		domain.TraceID("local-trace-"+spec.ID.String()),
		f.now,
		f.now.Add(2*time.Second),
	), nil
}

type fakeGraphProvider struct{}

func (fakeGraphProvider) GraphForTask(_ context.Context, task domain.TaskSpec) (codegraph.Graph, error) {
	store := codegraph.NewStore()
	fileID := codegraph.NodeID("file-" + task.ID.String())
	fnID := codegraph.NodeID("fn-" + task.ID.String())

	if err := store.AddNode(codegraph.NewFileNode(fileID, task.Oracle.GoldFiles[0])); err != nil {
		return nil, err
	}
	if err := store.AddNode(codegraph.NewFunctionNode(fnID, task.Oracle.GoldFiles[0], "score", 1, 10)); err != nil {
		return nil, err
	}
	if err := store.AddEdge(codegraph.Edge{
		From:   fileID,
		To:     fnID,
		Kind:   codegraph.EdgeContains,
		Weight: 1,
	}); err != nil {
		return nil, err
	}
	return store.Build()
}

type fakeScorer struct{}

func (fakeScorer) Score(_ context.Context, input score.Input) (score.ScoreSet, error) {
	if input.Run.Spec().System.Backend == domain.BackendIterativeContext {
		return score.NewScoreSet(
			score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 2},
			score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 3},
			score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.85},
			score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.15},
			score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.90},
		)
	}
	return score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 6},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 7},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.35},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.60},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.30},
	)
}

type fakeDecider struct{}

func (fakeDecider) Decide(comparisons []report.ScoreComparison, regressions []report.Regression) report.PromotionDecision {
	if len(regressions) > 0 {
		return report.PromotionDecision{
			Decision: report.DecisionReview,
			Reason:   "candidate has regressions in local fake comparison",
		}
	}
	for _, comparison := range comparisons {
		if comparison.Metric == score.MetricComposite && comparison.Candidate > comparison.Baseline {
			return report.PromotionDecision{
				Decision: report.DecisionPromote,
				Reason:   "candidate improves the composite score in local fake comparison",
			}
		}
	}
	return report.PromotionDecision{
		Decision: report.DecisionReview,
		Reason:   "candidate did not improve the composite score in local fake comparison",
	}
}
