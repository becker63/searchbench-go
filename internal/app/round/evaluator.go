package round

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"

	evaluatormodel "github.com/becker63/searchbench-go/internal/adapters/providers/evaluatormodel"
	langsmithtrace "github.com/becker63/searchbench-go/internal/adapters/trace/langsmith"
	evaluatoreino "github.com/becker63/searchbench-go/internal/agents/evaluator/eino"
	evaluatorcallbacks "github.com/becker63/searchbench-go/internal/agents/evaluator/eino/callbacks"
	evaluatorfake "github.com/becker63/searchbench-go/internal/agents/evaluator/fake"
	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
)

func runComparison(ctx context.Context, plan Plan, request evaluationRequest) (report.RoundReport, []EvaluatorExecution, error) {
	allowedTools := make(map[string]struct{}, len(plan.Evaluator.ToolPolicy.EffectiveAllowed))
	for _, name := range plan.Evaluator.ToolPolicy.EffectiveAllowed {
		allowedTools[name] = struct{}{}
	}

	modelFactory := request.EvaluatorModelFactory
	if modelFactory == nil {
		modelFactory = evaluatormodel.NewFactory(evaluatormodel.Config{
			Provider:        plan.Evaluator.Model.Provider,
			Model:           plan.Evaluator.Model.Name,
			MaxOutputTokens: plan.Evaluator.Model.MaxOutputTokens,
		})
	}

	executor := &evaluatorExecutor{
		modelFactory: modelFactory,
		toolFactory:  request.EvaluatorToolFactory,
		allowedTools: allowedTools,
		evaluatorAppendix: run.EvaluatorRunAppendix{
			SystemPrompt: plan.Evaluator.ToolPolicy.SystemPrompt,
		},
		retryPolicy: evaluatoreino.RetryPolicy{
			MaxAttempts:                plan.Evaluator.Retry.MaxAttempts,
			RetryOnModelError:          plan.Evaluator.Retry.RetryOnModelError,
			RetryOnToolFailure:         plan.Evaluator.Retry.RetryOnToolFailure,
			RetryOnFinalizationFailure: plan.Evaluator.Retry.RetryOnFinalizationFailure,
			RetryOnInvalidPrediction:   plan.Evaluator.Retry.RetryOnInvalidPrediction,
		},
	}

	runner := compare.Runner{
		Executor:      executor,
		GraphProvider: evaluatorfake.NewGraphProvider(),
		Scorer:        evaluatorfake.NewScorer(),
		Decider:       evaluatorfake.NewDecider(),
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
	allowedTools map[string]struct{}
	// evaluatorAppendix holds manifest-only evaluator text (e.g. systemPrompt).
	evaluatorAppendix run.EvaluatorRunAppendix
	retryPolicy       evaluatoreino.RetryPolicy

	mu      sync.Mutex
	records []EvaluatorExecution
}

func (e *evaluatorExecutor) Execute(ctx context.Context, spec run.Spec) (run.ExecutedRun, error) {
	spec = mergeEvaluatorAppendix(spec, e.evaluatorAppendix)

	modelFactory := e.modelFactory
	if modelFactory == nil {
		// Defensive: runComparison normally injects evaluatormodel.NewFactory.
		modelFactory = evaluatorfake.ModelFactory
	}
	model, err := modelFactory(spec)
	if err != nil {
		return run.ExecutedRun{}, fmt.Errorf("local evaluator model: %w", err)
	}

	toolFactory := e.toolFactory
	if toolFactory == nil {
		toolFactory = evaluatorfake.ToolFactory
	}
	tools, err := toolFactory(spec)
	if err != nil {
		return run.ExecutedRun{}, fmt.Errorf("local evaluator tools: %w", err)
	}
	filtered, ferr := filterToolsByAllowSet(ctx, tools, e.allowedTools)
	if ferr != nil {
		return run.ExecutedRun{}, ferr
	}
	tools = filtered

	var callbackFactories []evaluatorcallbacks.Factory
	langsmithFactory, lsErr := langsmithtrace.HandlerFactoryFromEnv()
	if lsErr != nil {
		return run.ExecutedRun{}, fmt.Errorf("langsmith trace callback: %w", lsErr)
	}
	if langsmithFactory != nil {
		callbackFactories = append(callbackFactories, evaluatorcallbacks.Factory(langsmithFactory))
	}

	evaluator, err := evaluatoreino.New(evaluatoreino.Config{
		Model:             model,
		Tools:             tools,
		WorkDir:           string(spec.Match.Repo.Path),
		RetryPolicy:       &e.retryPolicy,
		CallbackFactories: callbackFactories,
	})
	if err != nil {
		return run.ExecutedRun{}, fmt.Errorf("construct evaluator executor: %w", err)
	}

	runCtx := langsmithtrace.AugmentContext(ctx, langsmithtrace.ContextLabels{
		MatchID: spec.Match.ID.String(),
		RunID:   spec.ID.String(),
		Role:    string(roleForSpec(spec)),
	})

	result := evaluator.Run(runCtx, spec)
	e.recordExecution(spec, result)
	if result.Failure != nil {
		return run.ExecutedRun{}, result.Failure
	}
	return *result.Executed, nil
}

func (e *evaluatorExecutor) recordExecution(spec run.Spec, result evaluatoreino.Result) {
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

func mergeEvaluatorAppendix(spec run.Spec, appendix run.EvaluatorRunAppendix) run.Spec {
	if strings.TrimSpace(appendix.SystemPrompt) == "" {
		return spec
	}
	out := spec
	out.EvaluatorAppendix = run.EvaluatorRunAppendix{
		SystemPrompt: strings.TrimSpace(appendix.SystemPrompt),
	}
	return out
}

func filterToolsByAllowSet(ctx context.Context, tools []tool.BaseTool, allow map[string]struct{}) ([]tool.BaseTool, error) {
	if len(allow) == 0 {
		return tools, nil
	}
	var out []tool.BaseTool
	for i, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("local evaluator tools: tool %d info: %w", i, err)
		}
		if info == nil {
			continue
		}
		if _, ok := allow[info.Name]; ok {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("local evaluator tools: manifest tool policy excluded every registered tool")
	}
	return out, nil
}
