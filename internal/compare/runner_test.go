package compare

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/becker63/searchbench-go/internal/codegraph"
	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

var _ Executor = fakeExecutor{}
var _ GraphProvider = fakeGraphProvider{}
var _ Scorer = fakeScorer{}
var _ Decider = fakeDecider{}

func TestRunnerConvertsExecuteErrorToRunFailure(t *testing.T) {
	t.Parallel()

	runner := Runner{
		Executor: failingExecutor{
			err: errors.New("candidate execute failed"),
		},
		GraphProvider: fakeGraphProvider{},
		Scorer:        fakeScorer{},
		Decider:       fakeDecider{},
		NewRunID: func(role domain.Role, task domain.TaskSpec, system domain.SystemRef) domain.RunID {
			return domain.RunID(fmt.Sprintf("%s-%s-%s", role, task.ID, system.ID))
		},
		NewReportID: func() domain.ReportID { return domain.ReportID("report-2") },
	}

	got, err := runner.Run(context.Background(), NewPlan(
		domain.NewPair(
			exampleBaselineSystem(),
			exampleCandidateSystem("def score(task):\n    return 'candidate'\n"),
		),
		domain.NewNonEmpty(exampleTask(domain.TaskID("task-1"), domain.RepoRelPath("pkg/bug.go"))),
	))
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(got.Runs.Baseline) != 1 {
		t.Fatalf("len(got.Runs.Baseline) = %d, want 1", len(got.Runs.Baseline))
	}
	if len(got.Runs.Candidate) != 0 {
		t.Fatalf("len(got.Runs.Candidate) = %d, want 0", len(got.Runs.Candidate))
	}
	if len(got.Failures.Candidate) != 1 {
		t.Fatalf("len(got.Failures.Candidate) = %d, want 1", len(got.Failures.Candidate))
	}
	if got.Failures.Candidate[0].Stage != run.FailureExecute {
		t.Fatalf("candidate failure stage = %q, want %q", got.Failures.Candidate[0].Stage, run.FailureExecute)
	}
}

func TestRunnerParallelTasksPreservesPlanOrder(t *testing.T) {
	t.Parallel()

	runner := exampleRunner(fixedTestTime())
	runner.Executor = delayedExecutor{
		now: fixedTestTime(),
		delays: map[domain.TaskID]time.Duration{
			domain.TaskID("task-1"): 30 * time.Millisecond,
			domain.TaskID("task-2"): 10 * time.Millisecond,
			domain.TaskID("task-3"): 1 * time.Millisecond,
		},
	}
	runner.Parallelism = Parallelism{
		Mode:       ExecutionParallel,
		MaxWorkers: 2,
	}

	plan := NewPlan(
		domain.NewPair(
			exampleBaselineSystem(),
			exampleCandidateSystem("def score(task):\n    return 'candidate'\n"),
		),
		domain.NewNonEmpty(
			exampleTask(domain.TaskID("task-1"), domain.RepoRelPath("pkg/bug1.go")),
			exampleTask(domain.TaskID("task-2"), domain.RepoRelPath("pkg/bug2.go")),
			exampleTask(domain.TaskID("task-3"), domain.RepoRelPath("pkg/bug3.go")),
		),
	)

	got, err := runner.Run(context.Background(), plan)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	wantOrder := []domain.TaskID{
		domain.TaskID("task-1"),
		domain.TaskID("task-2"),
		domain.TaskID("task-3"),
	}
	assertRunOrder(t, got.Runs.Baseline, wantOrder)
	assertRunOrder(t, got.Runs.Candidate, wantOrder)
}

func TestRunnerParallelTasksRespectsContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := exampleRunner(fixedTestTime())
	runner.Parallelism = Parallelism{
		Mode:       ExecutionParallel,
		MaxWorkers: 2,
	}

	_, err := runner.Run(ctx, NewPlan(
		domain.NewPair(
			exampleBaselineSystem(),
			exampleCandidateSystem("def score(task):\n    return 'candidate'\n"),
		),
		domain.NewNonEmpty(
			exampleTask(domain.TaskID("task-1"), domain.RepoRelPath("pkg/bug1.go")),
			exampleTask(domain.TaskID("task-2"), domain.RepoRelPath("pkg/bug2.go")),
		),
	))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want %v", err, context.Canceled)
	}
}

type fakeExecutor struct {
	now time.Time
}

func (f fakeExecutor) Execute(_ context.Context, spec run.Spec) (run.ExecutedRun, error) {
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, domain.SessionID("session-"+spec.ID.String()))

	predictionFile := domain.RepoRelPath("pkg/baseline.go")
	reasoning := "baseline path"
	if spec.System.Backend == domain.BackendIterativeContext {
		predictionFile = domain.RepoRelPath("pkg/candidate.go")
		reasoning = "candidate path"
	}

	return run.NewExecuted(
		prepared,
		domain.Prediction{
			Files:     []domain.RepoRelPath{predictionFile},
			Reasoning: reasoning,
		},
		domain.UsageSummary{
			InputTokens:  100,
			OutputTokens: 20,
			TotalTokens:  120,
			CostUSD:      0.05,
		},
		domain.TraceID("trace-"+spec.ID.String()),
		f.now,
		f.now.Add(2*time.Second),
	), nil
}

type delayedExecutor struct {
	now    time.Time
	delays map[domain.TaskID]time.Duration
}

var _ Executor = delayedExecutor{}

func (d delayedExecutor) Execute(ctx context.Context, spec run.Spec) (run.ExecutedRun, error) {
	if delay := d.delays[spec.Task.ID]; delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			return run.ExecutedRun{}, ctx.Err()
		case <-timer.C:
		}
	}
	return fakeExecutor{now: d.now}.Execute(ctx, spec)
}

type failingExecutor struct {
	err error
}

var _ Executor = failingExecutor{}

func (f failingExecutor) Execute(ctx context.Context, spec run.Spec) (run.ExecutedRun, error) {
	if spec.System.Backend == domain.BackendIterativeContext {
		return run.ExecutedRun{}, f.err
	}
	return fakeExecutor{now: time.Now().UTC()}.Execute(ctx, spec)
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
			score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 1},
			score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 1},
			score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.9},
			score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.1},
			score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.95},
		)
	}
	return score.NewScoreSet(
		score.Metric[score.HopDistance]{Name: score.MetricGoldHop, Value: 4},
		score.Metric[score.HopDistance]{Name: score.MetricIssueHop, Value: 5},
		score.Metric[score.EfficiencyScore]{Name: score.MetricTokenEfficiency, Value: 0.4},
		score.Metric[score.CostScore]{Name: score.MetricCost, Value: 0.6},
		score.Metric[score.CompositeScore]{Name: score.MetricComposite, Value: 0.3},
	)
}

type fakeDecider struct{}

func (fakeDecider) Decide(_ []report.ScoreComparison, regressions []report.Regression) report.PromotionDecision {
	if len(regressions) == 0 {
		return report.PromotionDecision{
			Decision: report.DecisionPromote,
			Reason:   "candidate improved every required metric",
		}
	}
	return report.PromotionDecision{
		Decision: report.DecisionReview,
		Reason:   "candidate has regressions",
	}
}

func assertRunOrder(t *testing.T, runs []score.ScoredRun, want []domain.TaskID) {
	t.Helper()

	if len(runs) != len(want) {
		t.Fatalf("len(runs) = %d, want %d", len(runs), len(want))
	}
	for i, taskID := range want {
		if got := runs[i].Execution.Spec().Task.ID; got != taskID {
			t.Fatalf("runs[%d].Task.ID = %q, want %q", i, got, taskID)
		}
	}
}
