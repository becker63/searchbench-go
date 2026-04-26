package eino

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/domain"
	evaluatorprompt "github.com/becker63/searchbench-go/internal/prompts/evaluator"
	"github.com/becker63/searchbench-go/internal/run"
)

// RenderPromptFunc renders the evaluator prompt from a prompt-safe task view.
type RenderPromptFunc func(ctx context.Context, task domain.TaskSpec, allowedTools []string) (string, error)

// Config defines the minimal evaluator runner dependencies.
type Config struct {
	Model        model.ToolCallingChatModel
	Tools        []tool.BaseTool
	RenderPrompt RenderPromptFunc
	SessionID    domain.SessionID
	TraceID      domain.TraceID
	Now          func() time.Time
}

// Result is the typed outcome for one evaluator run.
type Result struct {
	Executed       *run.ExecutedRun
	Failure        *Failure
	Phases         []Phase
	RenderedPrompt string
	RawOutput      string
}

// Success reports whether the evaluator completed with a normalized
// prediction.
func (r Result) Success() bool {
	return r.Executed != nil && r.Failure == nil
}

// Evaluator is the minimal Eino-backed evaluator executor.
type Evaluator struct {
	model        model.ToolCallingChatModel
	tools        []tool.BaseTool
	renderPrompt RenderPromptFunc
	sessionID    domain.SessionID
	traceID      domain.TraceID
	now          func() time.Time
}

// New constructs a cold evaluator runner.
func New(config Config) (*Evaluator, error) {
	if config.Model == nil {
		return nil, fmt.Errorf("eino evaluator: model is required")
	}

	renderPrompt := config.RenderPrompt
	if renderPrompt == nil {
		renderPrompt = defaultRenderPrompt
	}

	now := config.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	sessionID := config.SessionID
	if sessionID.Empty() {
		sessionID = domain.SessionID("eino-local")
	}

	return &Evaluator{
		model:        config.Model,
		tools:        append([]tool.BaseTool(nil), config.Tools...),
		renderPrompt: renderPrompt,
		sessionID:    sessionID,
		traceID:      config.TraceID,
		now:          now,
	}, nil
}

// Execute implements the compare.Executor success path while preserving the
// richer typed failure detail through the returned error.
func (e *Evaluator) Execute(ctx context.Context, spec run.Spec) (run.ExecutedRun, error) {
	result := e.Run(ctx, spec)
	if result.Failure != nil {
		return run.ExecutedRun{}, result.Failure
	}
	return *result.Executed, nil
}

// Run executes the minimal evaluator path and returns a typed structured
// result.
func (e *Evaluator) Run(ctx context.Context, spec run.Spec) Result {
	result := Result{}
	recordPhase := func(phase Phase) {
		result.Phases = append(result.Phases, phase)
	}

	recordPhase(PhaseRenderPrompt)
	allowedTools, err := e.allowedToolNames(ctx)
	if err != nil {
		result.Failure = &Failure{
			Phase:   PhaseRenderPrompt,
			Kind:    FailureKindPromptRenderFailed,
			Message: "collect tool names",
			Cause:   err,
		}
		return result
	}

	renderedPrompt, err := e.renderPrompt(ctx, spec.Task, allowedTools)
	if err != nil {
		result.Failure = &Failure{
			Phase:   PhaseRenderPrompt,
			Kind:    FailureKindPromptRenderFailed,
			Message: "render evaluator prompt",
			Cause:   err,
		}
		return result
	}
	result.RenderedPrompt = renderedPrompt

	recordPhase(PhaseRunEvaluator)
	startedAt := e.now().UTC()
	rawOutput, toolErr, err := e.runEvaluator(ctx, renderedPrompt)
	if err != nil {
		kind := FailureKindEvaluatorFailed
		message := "run evaluator agent"
		cause := err
		if toolErr != nil {
			kind = FailureKindToolCallFailed
			message = "run evaluator tool call"
			cause = toolErr
		}
		result.Failure = &Failure{
			Phase:   PhaseRunEvaluator,
			Kind:    kind,
			Message: message,
			Cause:   cause,
		}
		return result
	}
	result.RawOutput = rawOutput

	recordPhase(PhaseFinalizePrediction)
	prediction, kind, err := FinalizePrediction(rawOutput)
	if err != nil {
		result.Failure = &Failure{
			Phase:   PhaseFinalizePrediction,
			Kind:    kind,
			Message: "finalize evaluator prediction",
			Cause:   err,
		}
		return result
	}

	recordPhase(PhaseComplete)
	planned := run.NewPlanned(spec)
	prepared := run.NewPrepared(planned, e.sessionID)
	executed := run.NewExecuted(
		prepared,
		prediction,
		domain.UsageSummary{},
		e.traceID,
		startedAt,
		e.now().UTC(),
	)
	result.Executed = &executed
	return result
}

func (e *Evaluator) allowedToolNames(ctx context.Context) ([]string, error) {
	if len(e.tools) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(e.tools))
	for i, t := range e.tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("tool %d info: %w", i, err)
		}
		names = append(names, info.Name)
	}
	return names, nil
}

func (e *Evaluator) runEvaluator(ctx context.Context, renderedPrompt string) (string, error, error) {
	recorder := &toolErrorRecorder{}
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "searchbench_evaluator",
		Description: "Minimal SearchBench evaluator agent",
		Model:       e.model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: e.tools,
				ToolCallMiddlewares: []compose.ToolMiddleware{
					{Invokable: recorder.WrapInvokable},
				},
			},
		},
	})
	if err != nil {
		return "", nil, err
	}

	iterator := agent.Run(ctx, &adk.AgentInput{
		Messages: []adk.Message{
			schema.UserMessage(renderedPrompt),
		},
	})

	var finalOutput string
	for {
		event, ok := iterator.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", recorder.err, event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		message, err := event.Output.MessageOutput.GetMessage()
		if err != nil {
			return "", recorder.err, err
		}

		if event.Output.MessageOutput.Role == schema.Assistant && len(message.ToolCalls) == 0 {
			finalOutput = message.Content
		}
	}

	return finalOutput, nil, nil
}

func defaultRenderPrompt(ctx context.Context, task domain.TaskSpec, allowedTools []string) (string, error) {
	return evaluatorprompt.Render(ctx, evaluatorprompt.InputFromTask(task, allowedTools))
}

type toolErrorRecorder struct {
	err error
}

func (r *toolErrorRecorder) WrapInvokable(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
	return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
		output, err := next(ctx, input)
		if err != nil && r.err == nil {
			r.err = fmt.Errorf("%s: %w", input.Name, err)
		}
		return output, err
	}
}
