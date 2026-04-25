package compare

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/becker63/searchbench-go/internal/codegraph"
	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/report"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/score"
)

// Executor is the orchestration seam for turning one planned run spec into a
// successful executed run without binding compare to a concrete runtime.
//
// Implementations used with parallel execution must be safe for concurrent
// calls. Parallel-safe executors should create isolated per-run state. They
// must not share mutable backend/session state across concurrent Execute calls
// unless they provide their own synchronization.
type Executor interface {
	Execute(ctx context.Context, spec run.Spec) (run.ExecutedRun, error)
}

// GraphProvider is the orchestration seam for loading a scoring graph for one
// task without binding compare to a concrete graph source.
//
// Implementations used with parallel execution must be safe for concurrent
// calls, or the runner must be configured with sequential execution.
type GraphProvider interface {
	GraphForTask(ctx context.Context, task domain.TaskSpec) (codegraph.Graph, error)
}

// Scorer is the orchestration seam for turning one executed run plus graph
// input into a complete required ScoreSet.
//
// Implementations used with parallel execution must be safe for concurrent
// calls. Parallel-safe scorers should not mutate shared scoring state without
// synchronization.
type Scorer interface {
	Score(ctx context.Context, input score.Input) (score.ScoreSet, error)
}

// Decider is the orchestration seam for reducing comparisons and regressions
// into a final promotion decision.
type Decider interface {
	Decide(comparisons []report.ScoreComparison, regressions []report.Regression) report.PromotionDecision
}

type RunIDFunc func(role domain.Role, task domain.TaskSpec, system domain.SystemRef) domain.RunID

type ReportIDFunc func() domain.ReportID

type ClockFunc func() time.Time

type Runner struct {
	Executor      Executor
	GraphProvider GraphProvider
	Scorer        Scorer
	Decider       Decider
	NewRunID      RunIDFunc
	NewReportID   ReportIDFunc
	Now           ClockFunc
	Parallelism   Parallelism
}

type TaskComparisonResult struct {
	Runs        domain.Pair[*score.ScoredRun]
	Failures    domain.Pair[*run.RunFailure]
	Regressions []report.Regression
}

func (r Runner) Run(ctx context.Context, plan Plan) (report.CandidateReport, error) {
	if err := plan.Validate(); err != nil {
		return report.CandidateReport{}, err
	}
	if err := r.Validate(); err != nil {
		return report.CandidateReport{}, err
	}

	taskResults, err := r.RunTasks(ctx, plan)
	if err != nil {
		return report.CandidateReport{}, err
	}

	results := NewResults(plan.Tasks.Len())
	for _, taskResult := range taskResults {
		results.AddTaskResult(taskResult.Result)
	}

	summary := results.Summary()
	decision := r.Decider.Decide(summary.Comparisons, summary.Regressions)
	out := report.NewCandidateReport(
		r.NewReportID(),
		plan.ReportSpec(),
		summary.Runs,
		summary.Failures,
		summary.Comparisons,
		summary.Regressions,
		decision,
	)
	if r.Now != nil {
		out.CreatedAt = r.Now().UTC()
	}
	return out, nil
}

func (r Runner) Validate() error {
	if r.Executor == nil {
		return errors.New("compare: executor is required")
	}
	if r.GraphProvider == nil {
		return errors.New("compare: graph provider is required")
	}
	if r.Scorer == nil {
		return errors.New("compare: scorer is required")
	}
	if r.Decider == nil {
		return errors.New("compare: decider is required")
	}
	if r.NewRunID == nil {
		return errors.New("compare: run id function is required")
	}
	if r.NewReportID == nil {
		return errors.New("compare: report id function is required")
	}
	if err := r.normalizedParallelism().Validate(); err != nil {
		return err
	}
	return nil
}

func (r Runner) normalizedParallelism() Parallelism {
	return r.Parallelism.Normalize()
}

func (r Runner) RunTasks(ctx context.Context, plan Plan) ([]TaskWorkResult, error) {
	tasks := plan.Tasks.All()
	parallelism := r.normalizedParallelism()

	if parallelism.Mode == ExecutionSequential {
		results := make([]TaskWorkResult, 0, len(tasks))
		for index, task := range tasks {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			result := r.CompareTask(ctx, plan.Systems, task)
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			results = append(results, TaskWorkResult{
				Index:  index,
				TaskID: task.ID,
				Result: result,
			})
		}
		return results, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type taskJob struct {
		index int
		task  domain.TaskSpec
	}

	type workerResult struct {
		work TaskWorkResult
		err  error
	}

	workerCount := parallelism.MaxWorkers
	if workerCount > len(tasks) {
		workerCount = len(tasks)
	}
	if workerCount == 0 {
		return nil, nil
	}

	jobs := make(chan taskJob)
	resultsCh := make(chan workerResult, len(tasks))

	for range workerCount {
		go func() {
			for job := range jobs {
				if err := ctx.Err(); err != nil {
					resultsCh <- workerResult{err: err}
					return
				}

				result := r.CompareTask(ctx, plan.Systems, job.task)
				if err := ctx.Err(); err != nil {
					resultsCh <- workerResult{err: err}
					return
				}

				resultsCh <- workerResult{
					work: TaskWorkResult{
						Index:  job.index,
						TaskID: job.task.ID,
						Result: result,
					},
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for index, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case jobs <- taskJob{index: index, task: task}:
			}
		}
	}()

	ordered := make([]TaskWorkResult, len(tasks))
	received := 0
	for received < len(tasks) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultsCh:
			if result.err != nil {
				if parallelism.FailFast {
					cancel()
				}
				return nil, result.err
			}
			ordered[result.work.Index] = result.work
			received++
		}
	}

	return ordered, nil
}

func (r Runner) CompareTask(
	ctx context.Context,
	systems domain.Pair[domain.SystemSpec],
	task domain.TaskSpec,
) TaskComparisonResult {
	type outcome struct {
		run     *score.ScoredRun
		failure *run.RunFailure
	}

	graph, err := r.GraphProvider.GraphForTask(ctx, task)
	if err != nil {
		return TaskComparisonResult{
			Failures: domain.MapPair(systems, func(role domain.Role, system domain.SystemSpec) *run.RunFailure {
				spec := run.NewSpec(r.NewRunID(role, task, system.Ref()), task, system)
				stageErr := NewStageError(spec, run.FailureScore, fmt.Errorf("graph: %w", err))
				failure := failureFromError(spec, run.FailureScore, stageErr)
				return &failure
			}),
		}
	}

	outcomes := domain.MapPair(systems, func(role domain.Role, system domain.SystemSpec) outcome {
		scoredRun, failure := r.ExecuteAndScore(ctx, role, task, system, graph)
		return outcome{
			run:     scoredRun,
			failure: failure,
		}
	})

	out := TaskComparisonResult{
		Runs: domain.NewPair(
			outcomes.Baseline.run,
			outcomes.Candidate.run,
		),
		Failures: domain.NewPair(
			outcomes.Baseline.failure,
			outcomes.Candidate.failure,
		),
	}
	if out.Runs.Baseline != nil && out.Runs.Candidate != nil {
		out.Regressions = report.RegressionsForTask(task.ID, score.CompareSets(out.Runs.Baseline.Scores, out.Runs.Candidate.Scores))
	}
	return out
}

func (r Runner) ExecuteAndScore(
	ctx context.Context,
	role domain.Role,
	task domain.TaskSpec,
	system domain.SystemSpec,
	graph codegraph.Graph,
) (*score.ScoredRun, *run.RunFailure) {
	spec := run.NewSpec(r.NewRunID(role, task, system.Ref()), task, system)

	executed, err := r.Executor.Execute(ctx, spec)
	if err != nil {
		stageErr := NewStageError(spec, run.FailureExecute, err)
		failure := failureFromError(spec, run.FailureExecute, stageErr)
		return nil, &failure
	}

	scoreSet, err := r.Scorer.Score(ctx, score.Input{
		Run:   executed,
		Graph: graph,
	})
	if err != nil {
		stageErr := NewStageError(spec, run.FailureScore, err)
		failure := failureFromError(spec, run.FailureScore, stageErr)
		return nil, &failure
	}

	scoredRun, err := score.NewScoredRun(executed, scoreSet)
	if err != nil {
		stageErr := NewStageError(spec, run.FailureScore, err)
		failure := failureFromError(spec, run.FailureScore, stageErr)
		return nil, &failure
	}

	return &scoredRun, nil
}
