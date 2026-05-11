package round

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	backendfake "github.com/becker63/searchbench-go/internal/adapters/backend/fake"
	codegraphfake "github.com/becker63/searchbench-go/internal/adapters/codegraph/fake"
	executoreino "github.com/becker63/searchbench-go/internal/adapters/executor/eino"
	scoringfake "github.com/becker63/searchbench-go/internal/adapters/scoring/fake"
	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

func runComparison(ctx context.Context, plan Plan, request evaluationRequest) (report.RoundReport, []EvaluatorExecution, error) {
	executor := &evaluatorExecutor{
		modelFactory: request.EvaluatorModelFactory,
		toolFactory:  request.EvaluatorToolFactory,
		retryPolicy: executoreino.RetryPolicy{
			MaxAttempts:                plan.Evaluator.Retry.MaxAttempts,
			RetryOnModelError:          plan.Evaluator.Retry.RetryOnModelError,
			RetryOnToolFailure:         plan.Evaluator.Retry.RetryOnToolFailure,
			RetryOnFinalizationFailure: plan.Evaluator.Retry.RetryOnFinalizationFailure,
			RetryOnInvalidPrediction:   plan.Evaluator.Retry.RetryOnInvalidPrediction,
		},
	}

	runner := compare.Runner{
		Executor:      executor,
		GraphProvider: codegraphfake.New(),
		Scorer:        scoringfake.New(),
		Decider:       backendfake.NewDecider(),
		NewRunID: func(role domain.Role, task domain.MatchSpec, system domain.SystemRef) domain.RunID {
			return domain.RunID(fmt.Sprintf("%s-%s-%s", role, task.ID, system.ID))
		},
		NewReportID: func() domain.ReportID {
			return plan.ReportID
		},
		Now: func() time.Time {
			return plan.CreatedAt
		},
		Parallelism: plan.Parallelism,
	}
	out, err := runner.Run(ctx, plan.ComparePlan())
	if err != nil {
		return report.RoundReport{}, nil, err
	}
	out.CreatedAt = plan.CreatedAt
	return out, executor.executions(), nil
}

type evaluatorExecutor struct {
	modelFactory EvaluatorModelFactory
	toolFactory  EvaluatorToolFactory
	retryPolicy  executoreino.RetryPolicy

	mu      sync.Mutex
	records []EvaluatorExecution
}

func (e *evaluatorExecutor) Execute(ctx context.Context, spec run.Spec) (run.ExecutedRun, error) {
	modelFactory := e.modelFactory
	if modelFactory == nil {
		modelFactory = backendfake.ModelFactory
	}
	model, err := modelFactory(spec)
	if err != nil {
		return run.ExecutedRun{}, fmt.Errorf("local evaluator model: %w", err)
	}

	toolFactory := e.toolFactory
	if toolFactory == nil {
		toolFactory = backendfake.ToolFactory
	}
	tools, err := toolFactory(spec)
	if err != nil {
		return run.ExecutedRun{}, fmt.Errorf("local evaluator tools: %w", err)
	}

	evaluator, err := executoreino.New(executoreino.Config{
		Model:       model,
		Tools:       tools,
		WorkDir:     string(spec.Match.Repo.Path),
		RetryPolicy: &e.retryPolicy,
	})
	if err != nil {
		return run.ExecutedRun{}, fmt.Errorf("construct evaluator executor: %w", err)
	}

	result := evaluator.Run(ctx, spec)
	e.recordExecution(spec, result)
	if result.Failure != nil {
		return run.ExecutedRun{}, result.Failure
	}
	return *result.Executed, nil
}

func (e *evaluatorExecutor) recordExecution(spec run.Spec, result executoreino.Result) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.records = append(e.records, EvaluatorExecution{
		Role:    roleForSpec(spec),
		MatchID: spec.Match.ID,
		RunID:   spec.ID,
		Result:  result,
	})
}

func (e *evaluatorExecutor) executions() []EvaluatorExecution {
	e.mu.Lock()
	defer e.mu.Unlock()

	out := make([]EvaluatorExecution, len(e.records))
	copy(out, e.records)
	return out
}

func roleForSpec(spec run.Spec) domain.Role {
	if strings.HasPrefix(spec.ID.String(), string(domain.RoleIncumbent)+"-") {
		return domain.RoleIncumbent
	}
	if strings.HasPrefix(spec.ID.String(), string(domain.RoleChallenger)+"-") {
		return domain.RoleChallenger
	}
	return ""
}
