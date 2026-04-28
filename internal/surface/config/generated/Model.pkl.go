// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/surface/config/generated/provider"

type Model struct {
	Provider provider.Provider `pkl:"provider"`

	Name string `pkl:"name"`

	Routing *string `pkl:"routing"`

	SystemPrompt *string `pkl:"systemPrompt"`

	MaxOutputTokens *int `pkl:"maxOutputTokens"`
}
