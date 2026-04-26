package eino

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/domain"
	"github.com/becker63/searchbench-go/internal/pipeline"
	evaluatorprompt "github.com/becker63/searchbench-go/internal/prompts/evaluator"
	"github.com/becker63/searchbench-go/internal/run"
)

// RenderPromptFunc renders the evaluator prompt from its typed prompt contract.
type RenderPromptFunc func(ctx context.Context, input evaluatorprompt.Input) (string, error)

// Config defines the minimal evaluator runner dependencies.
type Config struct {
	Model                    model.ToolCallingChatModel
	Tools                    []tool.BaseTool
	RenderPrompt             RenderPromptFunc
	CommandRunner            pipeline.CommandRunner
	PipelineSteps            []pipeline.CommandSpec
	Allowlist                pipeline.Allowlist
	WorkDir                  string
	RetryPolicy              *RetryPolicy
	SessionID                domain.SessionID
	TraceID                  domain.TraceID
	Now                      func() time.Time
	MaxPipelineFeedbackChars int
}

// Result is the typed outcome for one evaluator run.
type Result struct {
	Executed               *run.ExecutedRun
	Failure                *Failure
	Phases                 []Phase
	Attempts               []Attempt
	RenderedPrompt         string
	RawOutput              string
	PipelineResults        []pipeline.StepResult
	PipelineClassification *pipeline.Classification
}

// Success reports whether the evaluator completed with a normalized
// prediction.
func (r Result) Success() bool {
	return r.Executed != nil && r.Failure == nil
}

// Evaluator is the minimal Eino-backed evaluator executor.
type Evaluator struct {
	model                    model.ToolCallingChatModel
	tools                    []tool.BaseTool
	renderPrompt             RenderPromptFunc
	pipeline                 pipeline.Runner
	pipelineSteps            []pipeline.CommandSpec
	retryPolicy              RetryPolicy
	maxPipelineFeedbackChars int
	sessionID                domain.SessionID
	traceID                  domain.TraceID
	now                      func() time.Time
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

	workDir := config.WorkDir
	if workDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("eino evaluator: determine working directory: %w", err)
		}
		workDir = wd
	}

	pipelineSteps := append([]pipeline.CommandSpec(nil), config.PipelineSteps...)
	if len(pipelineSteps) == 0 {
		pipelineSteps = pipeline.DefaultEvaluatorSteps(workDir)
	}

	maxPipelineFeedbackChars := config.MaxPipelineFeedbackChars
	if maxPipelineFeedbackChars <= 0 {
		maxPipelineFeedbackChars = 2000
	}

	sessionID := config.SessionID
	if sessionID.Empty() {
		sessionID = domain.SessionID("eino-local")
	}

	return &Evaluator{
		model:        config.Model,
		tools:        append([]tool.BaseTool(nil), config.Tools...),
		renderPrompt: renderPrompt,
		pipeline: pipeline.Runner{
			CommandRunner: config.CommandRunner,
			Allowlist:     config.Allowlist,
		},
		pipelineSteps:            pipelineSteps,
		retryPolicy:              normalizeRetryPolicy(config.RetryPolicy),
		maxPipelineFeedbackChars: maxPipelineFeedbackChars,
		sessionID:                sessionID,
		traceID:                  config.TraceID,
		now:                      now,
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

	recordPhase(PhaseRunPipeline)
	pipelineResults := e.pipeline.Run(ctx, e.pipelineSteps)
	result.PipelineResults = append([]pipeline.StepResult(nil), pipelineResults...)

	classification := pipeline.Classify(pipelineResults)
	result.PipelineClassification = &classification
	if classification.HasFailures() {
		recordPhase(PhaseClassifyPipeline)
		kind := FailureKindPipelineFailed
		if len(classification.InfrastructureFailures) > 0 {
			kind = FailureKindPipelineInfrastructureFailed
		}
		result.Failure = &Failure{
			Phase:                  PhaseClassifyPipeline,
			Kind:                   kind,
			Message:                "pipeline validation failed",
			Recoverable:            false,
			StepResults:            append([]pipeline.StepResult(nil), pipelineResults...),
			PipelineClassification: result.PipelineClassification,
			PipelineFeedback:       pipeline.FormatPipelineFeedback(classification, e.maxPipelineFeedbackChars),
		}
		return result
	}

	allowedTools, err := e.allowedToolNames(ctx)
	if err != nil {
		failure := &Failure{
			Phase:   PhaseRenderPrompt,
			Kind:    FailureKindPromptRenderFailed,
			Message: "collect tool names",
			Cause:   err,
			Attempt: 1,
		}
		result.Attempts = append(result.Attempts, Attempt{Number: 1, Failure: failure})
		result.Failure = failure
		return result
	}

	for attemptNumber := 1; attemptNumber <= e.retryPolicy.MaxAttempts; attemptNumber++ {
		attempt := Attempt{Number: attemptNumber}

		recordPhase(PhaseRenderPrompt)
		input := evaluatorprompt.InputFromTask(spec.Task, allowedTools)
		for _, previous := range result.Attempts {
			if feedback := retryFeedbackForFailure(previous.Failure); feedback != "" {
				input.RetryFeedback = append(input.RetryFeedback, feedback)
			}
		}

		renderedPrompt, err := e.renderPrompt(ctx, input)
		attempt.RenderedPrompt = renderedPrompt
		result.RenderedPrompt = renderedPrompt
		if err != nil {
			failure := &Failure{
				Phase:   PhaseRenderPrompt,
				Kind:    FailureKindPromptRenderFailed,
				Message: "render evaluator prompt",
				Cause:   err,
				Attempt: attemptNumber,
			}
			attempt.Failure = failure
			result.Attempts = append(result.Attempts, attempt)
			result.Failure = failure
			return result
		}

		recordPhase(PhaseRunEvaluator)
		startedAt := e.now().UTC()
		rawOutput, toolErr, err := e.runEvaluator(ctx, renderedPrompt)
		attempt.RawOutput = rawOutput
		result.RawOutput = rawOutput
		if err != nil {
			failure := e.evaluatorFailure(attemptNumber, toolErr, err)
			attempt.Failure = failure
			result.Attempts = append(result.Attempts, attempt)
			if e.shouldRetry(failure, attemptNumber) {
				recordPhase(PhasePrepareRetry)
				continue
			}
			if failure.Recoverable && e.retryPolicy.MaxAttempts > 1 {
				recordPhase(PhaseExhausted)
				result.Failure = exhaustedFailure(failure, result.Attempts)
				return result
			}
			result.Failure = failure
			return result
		}

		recordPhase(PhaseFinalizePrediction)
		prediction, kind, err := FinalizePrediction(rawOutput)
		if err != nil {
			failure := &Failure{
				Phase:       PhaseFinalizePrediction,
				Kind:        kind,
				Message:     "finalize evaluator prediction",
				Cause:       err,
				Recoverable: e.retryPolicy.allows(kind),
				Attempt:     attemptNumber,
			}
			attempt.Failure = failure
			result.Attempts = append(result.Attempts, attempt)
			if e.shouldRetry(failure, attemptNumber) {
				recordPhase(PhasePrepareRetry)
				continue
			}
			if failure.Recoverable && e.retryPolicy.MaxAttempts > 1 {
				recordPhase(PhaseExhausted)
				result.Failure = exhaustedFailure(failure, result.Attempts)
				return result
			}
			result.Failure = failure
			return result
		}

		recordPhase(PhaseComplete)
		result.Attempts = append(result.Attempts, attempt)
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

	recordPhase(PhaseExhausted)
	result.Failure = exhaustedFailure(nil, result.Attempts)
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

func defaultRenderPrompt(ctx context.Context, input evaluatorprompt.Input) (string, error) {
	return evaluatorprompt.Render(ctx, input)
}

func (e *Evaluator) shouldRetry(failure *Failure, attemptNumber int) bool {
	return failure != nil && failure.Recoverable && attemptNumber < e.retryPolicy.MaxAttempts
}

func (e *Evaluator) evaluatorFailure(attemptNumber int, toolErr error, err error) *Failure {
	kind := FailureKindEvaluatorFailed
	message := "run evaluator agent"
	cause := err
	recoverable := isRecoverableModelError(err) && e.retryPolicy.allows(FailureKindEvaluatorFailed)
	if toolErr != nil {
		kind = FailureKindToolCallFailed
		message = "run evaluator tool call"
		cause = toolErr
		recoverable = isRecoverableToolError(toolErr) && e.retryPolicy.allows(FailureKindToolCallFailed)
	}

	return &Failure{
		Phase:       PhaseRunEvaluator,
		Kind:        kind,
		Message:     message,
		Cause:       cause,
		Recoverable: recoverable,
		Attempt:     attemptNumber,
	}
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
