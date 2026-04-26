package eino

// Phase is the evaluator-local lifecycle phase name.
type Phase string

const (
	PhaseRenderPrompt       Phase = "render_prompt"
	PhaseRunEvaluator       Phase = "run_evaluator"
	PhaseFinalizePrediction Phase = "finalize_prediction"
	PhaseComplete           Phase = "complete"
)

// FailureKind is the evaluator-local failure classification.
type FailureKind string

const (
	FailureKindPromptRenderFailed FailureKind = "prompt_render_failed"
	FailureKindEvaluatorFailed    FailureKind = "evaluator_failed"
	FailureKindToolCallFailed     FailureKind = "tool_call_failed"
	FailureKindFinalizationFailed FailureKind = "finalization_failed"
	FailureKindInvalidPrediction  FailureKind = "invalid_prediction"
)
