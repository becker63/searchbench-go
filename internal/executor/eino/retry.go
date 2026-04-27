package eino

import (
	"context"
	"errors"
	"fmt"
)

// RetryPolicy controls bounded evaluator retries for recoverable failures.
//
// These retries happen across evaluator attempts. They do not limit or count
// Eino-internal model turns or tool calls inside a single attempt.
type RetryPolicy struct {
	MaxAttempts                int
	RetryOnModelError          bool
	RetryOnToolFailure         bool
	RetryOnFinalizationFailure bool
	RetryOnInvalidPrediction   bool
}

// Attempt records one evaluator-local retry attempt outcome.
//
// An Attempt is a fresh evaluator pass over the task after a prior attempt
// failed in a retryable way. It is not a model turn and it is not a tool call.
type Attempt struct {
	Number         int
	RenderedPrompt string
	RawOutput      string
	Failure        *Failure
}

// DefaultRetryPolicy returns the minimal retry policy for the evaluator loop.
//
// The default policy retries evaluator-owned failures only. It does not
// continue partial Eino conversations and it does not cover writer-side
// validation pipelines.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:                2,
		RetryOnModelError:          true,
		RetryOnToolFailure:         false,
		RetryOnFinalizationFailure: true,
		RetryOnInvalidPrediction:   true,
	}
}

func normalizeRetryPolicy(policy *RetryPolicy) RetryPolicy {
	if policy == nil {
		return DefaultRetryPolicy()
	}

	normalized := *policy
	if normalized.MaxAttempts <= 0 {
		normalized.MaxAttempts = 1
	}
	return normalized
}

func (p RetryPolicy) allows(kind FailureKind) bool {
	switch kind {
	case FailureKindEvaluatorFailed:
		return p.RetryOnModelError
	case FailureKindToolCallFailed:
		return p.RetryOnToolFailure
	case FailureKindFinalizationFailed:
		return p.RetryOnFinalizationFailure
	case FailureKindInvalidPrediction:
		return p.RetryOnInvalidPrediction
	default:
		return false
	}
}

func isRecoverableModelError(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

type recoverableError interface {
	Recoverable() bool
}

func isRecoverableToolError(err error) bool {
	if err == nil {
		return false
	}
	var recoverable recoverableError
	return errors.As(err, &recoverable) && recoverable.Recoverable()
}

func retryFeedbackForFailure(failure *Failure) string {
	if failure == nil {
		return ""
	}

	switch failure.Kind {
	case FailureKindEvaluatorFailed:
		return fmt.Sprintf("Previous attempt failed during evaluator execution: %s.", safeFailureDetail(failure))
	case FailureKindToolCallFailed:
		return fmt.Sprintf("Previous attempt failed during a tool call: %s.", safeFailureDetail(failure))
	case FailureKindFinalizationFailed:
		return "Previous attempt returned malformed JSON."
	case FailureKindInvalidPrediction:
		return "Previous attempt returned empty predicted files."
	default:
		return ""
	}
}

func safeFailureDetail(failure *Failure) string {
	if failure == nil {
		return "unknown failure"
	}
	if failure.Cause != nil {
		return failure.Cause.Error()
	}
	if failure.Message != "" {
		return failure.Message
	}
	return string(failure.Kind)
}

func exhaustedFailure(last *Failure, attempts []Attempt) *Failure {
	failure := &Failure{
		Phase:       PhaseExhausted,
		Kind:        FailureKindRetriesExhausted,
		Message:     fmt.Sprintf("evaluator retries exhausted after %d attempt(s)", len(attempts)),
		Recoverable: false,
	}
	if last != nil {
		failure.Attempt = last.Attempt
		failure.Cause = last
	}
	return failure
}
