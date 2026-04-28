package eino

// Phase is the evaluator-local lifecycle phase name.
//
// These are harness phases, not Eino's internal model/tool turns. A single
// PhaseRunEvaluator span may include multiple model turns and tool calls before
// returning one final assistant payload to the harness.
//
// These names stay local to the Eino evaluator until more than one executor
// needs the same retry vocabulary.
type Phase string

const (
	PhaseRenderPrompt           Phase = "render_prompt"
	PhasePrepareCallbacks       Phase = "prepare_callbacks"
	PhasePrepareUsageAccounting Phase = "prepare_usage_accounting"
	PhaseRunEvaluator           Phase = "run_evaluator"
	PhaseFinalizeUsage          Phase = "finalize_usage"
	PhaseFinalizePrediction     Phase = "finalize_prediction"
	PhasePrepareRetry           Phase = "prepare_retry"
	PhaseExhausted              Phase = "exhausted"
	PhaseComplete               Phase = "complete"
)

// FailureKind is the evaluator-local failure classification.
//
// These kinds classify failures for one evaluator attempt. They do not model
// writer pipeline failures or individual Eino-internal turns.
type FailureKind string

const (
	FailureKindCallbackSetupFailed        FailureKind = "callback_setup_failed"
	FailureKindUsageAccountingSetupFailed FailureKind = "usage_accounting_setup_failed"
	FailureKindPromptRenderFailed         FailureKind = "prompt_render_failed"
	FailureKindEvaluatorFailed            FailureKind = "evaluator_failed"
	FailureKindToolCallFailed             FailureKind = "tool_call_failed"
	FailureKindFinalizationFailed         FailureKind = "finalization_failed"
	FailureKindInvalidPrediction          FailureKind = "invalid_prediction"
	FailureKindRetriesExhausted           FailureKind = "retries_exhausted"
	FailureKindUnexpectedInternal         FailureKind = "unexpected_internal_failure"
)
