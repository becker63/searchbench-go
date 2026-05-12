package eino

import (
	"context"
	"errors"
	"fmt"

	pureoptimizer "github.com/becker63/searchbench-go/internal/pure/optimizer"
)

func normalizeOptimizerRetryPolicy(policy *pureoptimizer.RetryPolicy) pureoptimizer.RetryPolicy {
	if policy == nil {
		return pureoptimizer.DefaultRetryPolicy()
	}
	normalized := *policy
	if normalized.MaxAttempts <= 0 {
		normalized.MaxAttempts = 1
	}
	return normalized
}

func isRecoverableModelError(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}

func retryFeedbackForOptimizerFailure(failure *pureoptimizer.Failure) string {
	if failure == nil {
		return ""
	}
	detail := failure.Message
	if failure.Cause != nil {
		detail = failure.Cause.Error()
	}

	switch failure.Kind {
	case pureoptimizer.FailureKindOptimizerFailed:
		return fmt.Sprintf("Previous attempt failed during optimizer execution: %s.", detail)
	case pureoptimizer.FailureKindNextChallengerFailed:
		return fmt.Sprintf("Previous attempt returned an invalid proposal: %s.", detail)
	case pureoptimizer.FailureKindPolicyPipelineFailed, pureoptimizer.FailureKindPolicyPipelineInfrastructure:
		if failure.PipelineFeedback != "" {
			return fmt.Sprintf("Previous attempt failed validation during %s: %s", failure.Phase, failure.PipelineFeedback)
		}
		return fmt.Sprintf("Previous attempt failed validation: %s.", detail)
	default:
		return ""
	}
}

func exhaustedOptimizerFailure(attempts []pureoptimizer.Attempt) *pureoptimizer.Failure {
	failure := &pureoptimizer.Failure{
		Phase:     pureoptimizer.PhaseExhausted,
		Kind:      pureoptimizer.FailureKindOptimizerRetriesExhausted,
		Message:   fmt.Sprintf("optimizer retries exhausted after %d attempt(s)", len(attempts)),
		Retryable: false,
	}
	if len(attempts) == 0 || attempts[len(attempts)-1].Failure == nil {
		return failure
	}
	last := attempts[len(attempts)-1].Failure
	failure.Attempt = last.Attempt
	failure.Cause = last
	failure.PipelineCategory = last.PipelineCategory
	failure.PipelineFeedback = last.PipelineFeedback
	return failure
}
