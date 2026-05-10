package eino

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/becker63/searchbench-go/internal/ports/pipeline"
	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
	optimizerprompt "github.com/becker63/searchbench-go/internal/pure/prompts/optimizer"
)

// RenderOptimizerPromptFunc renders the optimizer prompt from its typed prompt contract.
type RenderOptimizerPromptFunc func(ctx context.Context, input optimizerprompt.Input) (string, error)

// ProposalValidationResult is the typed output of staged policy validation.
type ProposalValidationResult struct {
	Results        []pipeline.StepResult
	Classification *pipeline.Classification
}

// ValidateProposalFunc stages and validates one policy proposal.
type ValidateProposalFunc func(ctx context.Context, proposal pureoptimizer.NextChallengerProposal) (ProposalValidationResult, *pureoptimizer.Failure)

// OptimizerConfig defines the first next-challenger runner dependencies.
type OptimizerConfig struct {
	Model            model.ToolCallingChatModel
	RenderPrompt     RenderOptimizerPromptFunc
	ValidateProposal ValidateProposalFunc
	RetryPolicy      *pureoptimizer.RetryPolicy
	WorkDir          string
}

// Optimizer is the minimal Eino-backed optimizer executor.
type Optimizer struct {
	model            model.ToolCallingChatModel
	renderPrompt     RenderOptimizerPromptFunc
	validateProposal ValidateProposalFunc
	retryPolicy      pureoptimizer.RetryPolicy
}

// NewOptimizer constructs a cold next-challenger runner.
func NewOptimizer(config OptimizerConfig) (*Optimizer, error) {
	if config.Model == nil {
		return nil, fmt.Errorf("eino optimizer: model is required")
	}
	if config.ValidateProposal == nil {
		return nil, fmt.Errorf("eino optimizer: proposal validator is required")
	}
	renderPrompt := config.RenderPrompt
	if renderPrompt == nil {
		renderPrompt = defaultRenderOptimizerPrompt
	}
	if config.WorkDir == "" {
		if _, err := os.Getwd(); err != nil {
			return nil, fmt.Errorf("eino optimizer: determine working directory: %w", err)
		}
	}

	policy := pureoptimizer.DefaultRetryPolicy()
	if config.RetryPolicy != nil {
		policy = normalizeOptimizerRetryPolicy(config.RetryPolicy)
	}

	return &Optimizer{
		model:            config.Model,
		renderPrompt:     renderPrompt,
		validateProposal: config.ValidateProposal,
		retryPolicy:      policy,
	}, nil
}

// Run executes one optimizer attempt sequence and returns a typed result.
func (o *Optimizer) Run(ctx context.Context, spec pureoptimizer.Spec) pureoptimizer.NextChallengerRecord {
	result := pureoptimizer.NextChallengerRecord{}
	recordPhase := func(phase pureoptimizer.Phase) {
		result.Phases = append(result.Phases, phase)
	}

	for attemptNumber := 1; attemptNumber <= o.retryPolicy.MaxAttempts; attemptNumber++ {
		attempt := pureoptimizer.Attempt{
			Number: attemptNumber,
			State:  pureoptimizer.AttemptStatePending,
		}

		recordPhase(pureoptimizer.PhaseRenderOptimizerPrompt)
		input, err := optimizerprompt.InputFromSpec(spec)
		if err != nil {
			failure := &pureoptimizer.Failure{
				Phase:   pureoptimizer.PhaseRenderOptimizerPrompt,
				Kind:    pureoptimizer.FailureKindOptimizerPromptFailed,
				Message: "build optimizer prompt input",
				Cause:   err,
				Attempt: attemptNumber,
			}
			attempt.Failure = failure
			attempt.State = pureoptimizer.AttemptStateFailed
			result.Attempts = append(result.Attempts, attempt)
			result.Failure = failure
			return result
		}
		for _, previous := range result.Attempts {
			if feedback := retryFeedbackForOptimizerFailure(previous.Failure); feedback != "" {
				input.RetryFeedback = append(input.RetryFeedback, feedback)
				attempt.RetryFeedback = feedback
			}
		}

		renderedPrompt, err := o.renderPrompt(ctx, input)
		attempt.RenderedPrompt = renderedPrompt
		attempt.State = pureoptimizer.AttemptStatePromptRendered
		result.RenderedPrompt = renderedPrompt
		if err != nil {
			failure := &pureoptimizer.Failure{
				Phase:   pureoptimizer.PhaseRenderOptimizerPrompt,
				Kind:    pureoptimizer.FailureKindOptimizerPromptFailed,
				Message: "render optimizer prompt",
				Cause:   err,
				Attempt: attemptNumber,
			}
			attempt.Failure = failure
			attempt.State = pureoptimizer.AttemptStateFailed
			result.Attempts = append(result.Attempts, attempt)
			result.Failure = failure
			return result
		}

		recordPhase(pureoptimizer.PhaseRunOptimizer)
		message, failure := o.runOptimizer(ctx, renderedPrompt, attemptNumber)
		if failure != nil {
			attempt.Failure = failure
			attempt.State = pureoptimizer.AttemptStateFailed
			result.Attempts = append(result.Attempts, attempt)
			if failure.Retryable && attemptNumber < o.retryPolicy.MaxAttempts {
				recordPhase(pureoptimizer.PhasePrepareRetry)
				continue
			}
			result.Failure = failure
			return result
		}
		if message != nil {
			attempt.RawOutput = strings.TrimSpace(message.Content)
			result.RawOutput = attempt.RawOutput
		}

		recordPhase(pureoptimizer.PhaseFinalizeNextChallenger)
		proposal, failure := finalizeProposal(attempt.RawOutput, spec.Target, attemptNumber)
		if failure != nil {
			attempt.Failure = failure
			attempt.State = pureoptimizer.AttemptStateFailed
			result.Attempts = append(result.Attempts, attempt)
			if failure.Retryable && attemptNumber < o.retryPolicy.MaxAttempts {
				recordPhase(pureoptimizer.PhasePrepareRetry)
				continue
			}
			result.Failure = failure
			return result
		}
		attempt.Proposal = proposal
		attempt.State = pureoptimizer.AttemptStateNextChallengerFinalized

		recordPhase(pureoptimizer.PhaseWriteNextChallengerStage)
		recordPhase(pureoptimizer.PhaseRunPolicyPipeline)
		validation, failure := o.validateProposal(ctx, *proposal)
		attempt.PipelineResults = validation.stepResults()
		attempt.PipelineClassification = validation.classification()
		recordPhase(pureoptimizer.PhaseClassifyPipeline)
		if failure != nil {
			attempt.Failure = failure
			attempt.State = pureoptimizer.AttemptStatePipelineFailed
			result.Attempts = append(result.Attempts, attempt)
			if failure.Retryable && attemptNumber < o.retryPolicy.MaxAttempts {
				recordPhase(pureoptimizer.PhasePrepareRetry)
				continue
			}
			result.Failure = failure
			return result
		}

		recordPhase(pureoptimizer.PhaseAcceptNextChallenger)
		recordPhase(pureoptimizer.PhaseComplete)
		attempt.State = pureoptimizer.AttemptStateAccepted
		result.Attempts = append(result.Attempts, attempt)
		result.Success = true
		result.Proposal = proposal
		return result
	}

	recordPhase(pureoptimizer.PhaseExhausted)
	result.Failure = exhaustedOptimizerFailure(result.Attempts)
	return result
}

func (o *Optimizer) runOptimizer(ctx context.Context, renderedPrompt string, attemptNumber int) (*schema.Message, *pureoptimizer.Failure) {
	if err := ctx.Err(); err != nil {
		return nil, contextCancelledFailure(err, attemptNumber)
	}

	messages := []*schema.Message{
		schema.SystemMessage("You are the SearchBench optimizer."),
		schema.UserMessage(renderedPrompt),
	}
	message, err := o.model.Generate(ctx, messages)
	if err != nil {
		failure := &pureoptimizer.Failure{
			Phase:     pureoptimizer.PhaseRunOptimizer,
			Kind:      pureoptimizer.FailureKindOptimizerFailed,
			Message:   "run optimizer model",
			Cause:     err,
			Attempt:   attemptNumber,
			Retryable: o.retryPolicy.RetryOnModelError && isRecoverableModelError(err),
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, contextCancelledFailure(err, attemptNumber)
		}
		return nil, failure
	}
	if message == nil {
		return nil, &pureoptimizer.Failure{
			Phase:     pureoptimizer.PhaseRunOptimizer,
			Kind:      pureoptimizer.FailureKindOptimizerFailed,
			Message:   "optimizer returned no message",
			Attempt:   attemptNumber,
			Retryable: o.retryPolicy.RetryOnModelError,
		}
	}
	return message, nil
}

func defaultRenderOptimizerPrompt(ctx context.Context, input optimizerprompt.Input) (string, error) {
	return optimizerprompt.Render(ctx, input)
}

func contextCancelledFailure(err error, attemptNumber int) *pureoptimizer.Failure {
	return &pureoptimizer.Failure{
		Phase:     pureoptimizer.PhaseRunOptimizer,
		Kind:      pureoptimizer.FailureKindContextCancelled,
		Message:   "optimizer context cancelled",
		Cause:     err,
		Attempt:   attemptNumber,
		Retryable: false,
	}
}

func (r ProposalValidationResult) stepResults() []pipeline.StepResult {
	return append([]pipeline.StepResult(nil), r.Results...)
}

func (r ProposalValidationResult) classification() *pipeline.Classification {
	if r.Classification == nil {
		return nil
	}
	classification := *r.Classification
	return &classification
}
