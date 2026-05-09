// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

type RetryPolicy struct {
	MaxAttempts int `pkl:"maxAttempts"`

	RetryOnModelError bool `pkl:"retryOnModelError"`

	RetryOnToolFailure bool `pkl:"retryOnToolFailure"`

	RetryOnFinalizationFailure bool `pkl:"retryOnFinalizationFailure"`

	RetryOnInvalidPrediction bool `pkl:"retryOnInvalidPrediction"`
}
