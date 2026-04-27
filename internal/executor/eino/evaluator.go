package eino

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino/adk"
	einocallbacks "github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/domain"
	evaluatorcallbacks "github.com/becker63/searchbench-go/internal/executor/eino/callbacks"
	evaluatorprompt "github.com/becker63/searchbench-go/internal/prompts/evaluator"
	"github.com/becker63/searchbench-go/internal/run"
	"github.com/becker63/searchbench-go/internal/usage"
)

// RenderPromptFunc renders the evaluator prompt from its typed prompt contract.
type RenderPromptFunc func(ctx context.Context, input evaluatorprompt.Input) (string, error)

// UsageCollectorFactory constructs one run-local usage collector.
type UsageCollectorFactory func(spec run.Spec) (*usage.Collector, error)

const defaultAgentMaxIterations = 20

// Config defines the minimal evaluator runner dependencies.
//
// The harness owns evaluator-level bounds such as retry count and context
// cancellation. Lower-level bounds such as max model turns, tool-call limits,
// and token budgets belong to the configured Eino model/agent runtime rather
// than to evaluator business logic here. The evaluator maps
// spec.System.Runtime.MaxSteps onto Eino's MaxIterations bound when provided.
//
// CallbackFactories are optional per-attempt callback constructors. They are
// composed through the sibling callbacks package and must remain cold.
type Config struct {
	Model                  model.ToolCallingChatModel
	Tools                  []tool.BaseTool
	CallbackFactories      []evaluatorcallbacks.Factory
	UsageCollectorFactory  UsageCollectorFactory
	DisableUsageAccounting bool
	RenderPrompt           RenderPromptFunc
	WorkDir                string
	RetryPolicy            *RetryPolicy
	SessionID              domain.SessionID
	TraceID                domain.TraceID
	Now                    func() time.Time
}

// Result is the typed outcome for one evaluator run.
//
// A run is one bounded attempt to solve one task. The underlying Eino agent
// may take multiple model turns and tool calls during that run, but the result
// always captures one final prediction or one typed failure.
type Result struct {
	Executed       *run.ExecutedRun
	Failure        *Failure
	Phases         []Phase
	Attempts       []Attempt
	RenderedPrompt string
	RawOutput      string
	UsageRecords   []usage.Record
	UsageSummary   usage.Summary
}

// Success reports whether the evaluator completed with a normalized
// prediction.
func (r Result) Success() bool {
	return r.Executed != nil && r.Failure == nil
}

// Evaluator is the minimal Eino-backed evaluator executor.
//
// It delegates the internal model/tool loop to Eino and keeps only harness
// concerns local: prompt rendering, retry boundaries, finalization, and typed
// run results.
type Evaluator struct {
	model                  model.ToolCallingChatModel
	tools                  []tool.BaseTool
	callbackFactories      []evaluatorcallbacks.Factory
	usageCollectorFactory  UsageCollectorFactory
	disableUsageAccounting bool
	renderPrompt           RenderPromptFunc
	retryPolicy            RetryPolicy
	sessionID              domain.SessionID
	traceID                domain.TraceID
	now                    func() time.Time
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

	sessionID := config.SessionID
	if sessionID.Empty() {
		sessionID = domain.SessionID("eino-local")
	}

	usageCollectorFactory := config.UsageCollectorFactory
	if usageCollectorFactory == nil {
		usageCollectorFactory = func(spec run.Spec) (*usage.Collector, error) {
			return usage.NewCollector(usage.Config{
				DefaultProvider: spec.System.Model.Provider,
				DefaultModel:    spec.System.Model.Name,
			})
		}
	}

	return &Evaluator{
		model:                  config.Model,
		tools:                  append([]tool.BaseTool(nil), config.Tools...),
		callbackFactories:      append([]evaluatorcallbacks.Factory(nil), config.CallbackFactories...),
		usageCollectorFactory:  usageCollectorFactory,
		disableUsageAccounting: config.DisableUsageAccounting,
		renderPrompt:           renderPrompt,
		retryPolicy:            normalizeRetryPolicy(config.RetryPolicy),
		sessionID:              sessionID,
		traceID:                config.TraceID,
		now:                    now,
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

// Run executes one bounded evaluator attempt sequence for one task and returns
// a typed structured result.
//
// Each evaluator attempt renders a fresh prompt, lets Eino drive its internal
// model/tool loop until it returns a final assistant message or an error, then
// finalizes exactly one prediction. Callback construction happens in a separate
// prepare_callbacks phase before Eino execution begins. Retry attempts are new
// evaluator attempts, not continuations of earlier model/tool turns.
//
// Run does not execute CLI validation or writer repair pipeline behavior.
func (e *Evaluator) Run(ctx context.Context, spec run.Spec) Result {
	result := Result{}
	recordPhase := func(phase Phase) {
		result.Phases = append(result.Phases, phase)
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

		recordPhase(PhasePrepareCallbacks)
		callbackHandlers, err := evaluatorcallbacks.BuildCallbacks(ctx, evaluatorcallbacks.Config{
			Factories: e.callbackFactories,
		})
		if err != nil {
			failure := &Failure{
				Phase:   PhasePrepareCallbacks,
				Kind:    FailureKindCallbackSetupFailed,
				Message: "build evaluator callbacks",
				Cause:   err,
				Attempt: attemptNumber,
			}
			attempt.Failure = failure
			result.Attempts = append(result.Attempts, attempt)
			result.Failure = failure
			return result
		}

		var usageCallback *evaluatorcallbacks.UsageCallback
		if !e.disableUsageAccounting {
			usageCallback, err = evaluatorcallbacks.NewUsageCallback(evaluatorcallbacks.UsageConfig{
				Phase:           string(PhaseRunEvaluator),
				DefaultProvider: spec.System.Model.Provider,
				DefaultModel:    spec.System.Model.Name,
			})
			if err != nil {
				failure := &Failure{
					Phase:   PhasePrepareCallbacks,
					Kind:    FailureKindCallbackSetupFailed,
					Message: "build usage callback",
					Cause:   err,
					Attempt: attemptNumber,
				}
				attempt.Failure = failure
				result.Attempts = append(result.Attempts, attempt)
				result.Failure = failure
				return result
			}
			callbackHandlers = append(callbackHandlers, usageCallback.Handler())
		}

		var usageCollector *usage.Collector
		if usageCallback != nil {
			recordPhase(PhasePrepareUsageAccounting)
			usageCollector, err = e.usageCollectorFactory(spec)
			if err != nil {
				failure := &Failure{
					Phase:   PhasePrepareUsageAccounting,
					Kind:    FailureKindUsageAccountingSetupFailed,
					Message: "create usage collector",
					Cause:   err,
					Attempt: attemptNumber,
				}
				attempt.Failure = failure
				result.Attempts = append(result.Attempts, attempt)
				result.Failure = failure
				return result
			}
			if err := usageCallback.AttachCollector(usageCollector); err != nil {
				failure := &Failure{
					Phase:   PhasePrepareUsageAccounting,
					Kind:    FailureKindUsageAccountingSetupFailed,
					Message: "attach usage collector",
					Cause:   err,
					Attempt: attemptNumber,
				}
				attempt.Failure = failure
				result.Attempts = append(result.Attempts, attempt)
				result.Failure = failure
				return result
			}
		}

		recordPhase(PhaseRunEvaluator)
		startedAt := e.now().UTC()
		rawOutput, toolErr, err := e.runEvaluator(ctx, renderedPrompt, maxIterationsForSpec(spec), callbackHandlers)
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

		if usageCollector != nil {
			recordPhase(PhaseFinalizeUsage)
			result.UsageRecords = usageCollector.Records()
			result.UsageSummary = usageCollector.Summary()
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
			result.UsageSummary.DomainSummary(),
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

func (e *Evaluator) runEvaluator(ctx context.Context, renderedPrompt string, maxIterations int, callbackHandlers []einocallbacks.Handler) (string, error, error) {
	recorder := &toolErrorRecorder{}
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "searchbench_evaluator",
		Description:   "Minimal SearchBench evaluator agent",
		Model:         e.model,
		MaxIterations: maxIterations,
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

	runOptions := make([]adk.AgentRunOption, 0, 1)
	if len(callbackHandlers) > 0 {
		runOptions = append(runOptions, adk.WithCallbacks(callbackHandlers...))
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent})
	iterator := runner.Run(ctx, []adk.Message{
		schema.UserMessage(renderedPrompt),
	}, runOptions...)

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

		// Eino may emit multiple assistant/tool turns inside one evaluator run.
		// The harness keeps only the final assistant message without tool calls as
		// the candidate prediction payload for strict finalization.
		if event.Output.MessageOutput.Role == schema.Assistant && len(message.ToolCalls) == 0 {
			finalOutput = message.Content
		}
	}

	return finalOutput, nil, nil
}

func defaultRenderPrompt(ctx context.Context, input evaluatorprompt.Input) (string, error) {
	return evaluatorprompt.Render(ctx, input)
}

func maxIterationsForSpec(spec run.Spec) int {
	if spec.System.Runtime.MaxSteps > 0 {
		return spec.System.Runtime.MaxSteps
	}
	return defaultAgentMaxIterations
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
