package round

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	executoreino "github.com/becker63/searchbench-go/internal/adapters/executor/eino"
	"github.com/becker63/searchbench-go/internal/app/round/internal/compare"
	"github.com/becker63/searchbench-go/internal/pure/codegraph"
	"github.com/becker63/searchbench-go/internal/pure/domain"
	run "github.com/becker63/searchbench-go/internal/pure/execution"
	"github.com/becker63/searchbench-go/internal/pure/report"
	"github.com/becker63/searchbench-go/internal/pure/score"
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
		GraphProvider: fakeGraphProvider{},
		Scorer:        fakeScorer{},
		Decider:       fakeDecider{},
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
		modelFactory = defaultEvaluatorModelFactory
	}
	model, err := modelFactory(spec)
	if err != nil {
		return run.ExecutedRun{}, fmt.Errorf("local evaluator model: %w", err)
	}

	toolFactory := e.toolFactory
	if toolFactory == nil {
		toolFactory = defaultEvaluatorToolFactory
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

func defaultEvaluatorModelFactory(spec run.Spec) (model.ToolCallingChatModel, error) {
	return &defaultFakeEvaluatorModel{spec: spec}, nil
}

func defaultEvaluatorToolFactory(spec run.Spec) ([]tool.BaseTool, error) {
	return []tool.BaseTool{
		fakeRepoTool{
			spec: spec,
			name: "resolve",
			desc: "Return the most plausible fake repository paths for the issue.",
			run: func(spec run.Spec, _ map[string]any) any {
				return map[string]any{
					"paths": []string{"src/search_target.go", "src/incumbent_guess.go"},
					"issue": spec.Match.Input.Title,
				}
			},
		},
		fakeRepoTool{
			spec: spec,
			name: "expand",
			desc: "Return a deterministic fake file snippet for the requested path.",
			run: func(spec run.Spec, args map[string]any) any {
				path, _ := args["path"].(string)
				if strings.TrimSpace(path) == "" {
					path = "src/search_target.go"
				}
				return map[string]any{
					"path":    path,
					"snippet": fakeSnippetForPath(path),
				}
			},
		},
		fakeRepoTool{
			spec: spec,
			name: "resolve_and_expand",
			desc: "Resolve the likely file set and return one fake structural snippet.",
			run: func(spec run.Spec, _ map[string]any) any {
				return map[string]any{
					"paths": []string{"src/search_target.go", "src/incumbent_guess.go"},
					"files": []map[string]string{
						{
							"path":    "src/search_target.go",
							"snippet": fakeSnippetForPath("src/search_target.go"),
						},
						{
							"path":    "src/incumbent_guess.go",
							"snippet": fakeSnippetForPath("src/incumbent_guess.go"),
						},
					},
					"repo": string(spec.Match.Repo.Name),
				}
			},
		},
	}, nil
}

type defaultFakeEvaluatorModel struct {
	spec  run.Spec
	tools []*schema.ToolInfo
}

func (m *defaultFakeEvaluatorModel) Generate(ctx context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if hasToolResponse(input) {
		return schema.AssistantMessage(finalPredictionJSON(m.spec), nil), nil
	}

	toolName := "resolve_and_expand"
	if len(m.tools) > 0 {
		toolName = m.tools[0].Name
		for _, info := range m.tools {
			if info != nil && info.Name == "resolve_and_expand" {
				toolName = info.Name
				break
			}
		}
	}

	args := map[string]string{
		"query": m.spec.Match.Input.Title,
	}
	rawArgs, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	return schema.AssistantMessage("", []schema.ToolCall{{
		ID: "call-resolve-and-expand",
		Function: schema.FunctionCall{
			Name:      toolName,
			Arguments: string(rawArgs),
		},
	}}), nil
}

func (m *defaultFakeEvaluatorModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	message, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{message}), nil
}

func (m *defaultFakeEvaluatorModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return &defaultFakeEvaluatorModel{
		spec:  m.spec,
		tools: cloneToolInfos(tools),
	}, nil
}

type fakeRepoTool struct {
	spec run.Spec
	name string
	desc string
	run  func(run.Spec, map[string]any) any
}

func (t fakeRepoTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.name,
		Desc: t.desc,
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"query": {
				Desc: "Issue or symbol hint to search for.",
				Type: schema.String,
			},
			"path": {
				Desc: "Repository-relative path to expand.",
				Type: schema.String,
			},
		}),
	}, nil
}

func (t fakeRepoTool) InvokableRun(_ context.Context, input string, _ ...tool.Option) (string, error) {
	var args map[string]any
	if strings.TrimSpace(input) != "" {
		if err := json.Unmarshal([]byte(input), &args); err != nil {
			return "", fmt.Errorf("%s arguments: %w", t.name, err)
		}
	}
	raw, err := json.Marshal(t.run(t.spec, args))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

var _ tool.InvokableTool = fakeRepoTool{}

func hasToolResponse(messages []*schema.Message) bool {
	for _, message := range messages {
		if message != nil && message.Role == schema.Tool {
			return true
		}
	}
	return false
}

func finalPredictionJSON(spec run.Spec) string {
	path := "src/incumbent_guess.go"
	reasoning := "incumbent stayed conservative after the fake structural lookup"
	if spec.System.Backend == domain.BackendIterativeContext {
		path = string(spec.Match.Oracle.GoldFiles[0])
		reasoning = "challenger used the fake structural lookup to narrow onto the search target"
	}

	payload := struct {
		PredictedFiles []string `json:"predicted_files"`
		Reasoning      string   `json:"reasoning"`
	}{
		PredictedFiles: []string{path},
		Reasoning:      reasoning,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf(`{"predicted_files":["%s"]}`, path)
	}
	return string(data)
}

func fakeSnippetForPath(path string) string {
	switch filepath := strings.TrimSpace(path); filepath {
	case "src/search_target.go":
		return "func locateRetryTarget() { /* deterministic fake target */ }"
	case "src/incumbent_guess.go":
		return "func incumbentGuess() { /* deterministic fake fallback */ }"
	default:
		return "func fakeSnippet() {}"
	}
}

func cloneToolInfos(tools []*schema.ToolInfo) []*schema.ToolInfo {
	if len(tools) == 0 {
		return nil
	}

	cloned := make([]*schema.ToolInfo, len(tools))
	for i, info := range tools {
		if info == nil {
			continue
		}
		copyInfo := *info
		cloned[i] = &copyInfo
	}
	return cloned
}

type fakeGraphProvider struct{}

func (fakeGraphProvider) GraphForTask(_ context.Context, task domain.MatchSpec) (codegraph.Graph, error) {
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

func (fakeDecider) Decide(comparisons []report.ScoreComparison, regressions []report.Regression) report.Decision {
	if len(regressions) > 0 {
		return report.Decision{
			Decision: report.DecisionReview,
			Reason:   "challenger has regressions in local fake comparison",
		}
	}
	for _, comparison := range comparisons {
		if comparison.Metric == score.MetricComposite && comparison.Challenger > comparison.Incumbent {
			return report.Decision{
				Decision: report.DecisionPromoteChallenger,
				Reason:   "challenger improves the composite score in local fake comparison",
			}
		}
	}
	return report.Decision{
		Decision: report.DecisionReview,
		Reason:   "challenger did not improve the composite score in local fake comparison",
	}
}
