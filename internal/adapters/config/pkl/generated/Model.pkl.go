// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/provider"

// LLM call routing for an agent’s chat/completion model.
type Model struct {
	// Provider key consumed by the harness / adapter.
	Provider provider.Provider `pkl:"provider"`

	// Model name within the provider.
	Name string `pkl:"name"`

	// Optional provider-specific routing hint (e.g. deployment id).
	Routing *string `pkl:"routing"`

	// Optional max output tokens for the model call (positive when set).
	MaxOutputTokens *int `pkl:"maxOutputTokens"`
}
