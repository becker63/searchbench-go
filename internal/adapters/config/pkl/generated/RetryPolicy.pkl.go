// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// How the harness retries failed evaluator attempts before surfacing a run failure.
type RetryPolicy struct {
	// Upper bound on distinct evaluator attempts including retries (must be positive).
	MaxAttempts int `pkl:"maxAttempts"`

	// Whether to retry after provider/model transport style errors.
	RetryOnModelError bool `pkl:"retryOnModelError"`

	// Whether to retry after tool execution failures inside the evaluator agent loop.
	RetryOnToolFailure bool `pkl:"retryOnToolFailure"`

	// Whether to retry when JSON prediction finalization fails.
	RetryOnFinalizationFailure bool `pkl:"retryOnFinalizationFailure"`

	// Whether to retry when the model returns a prediction that fails schema validation.
	RetryOnInvalidPrediction bool `pkl:"retryOnInvalidPrediction"`
}
