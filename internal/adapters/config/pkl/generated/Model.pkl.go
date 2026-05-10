// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/provider"

type Model struct {
	Provider provider.Provider `pkl:"provider"`

	Name string `pkl:"name"`

	Routing *string `pkl:"routing"`

	MaxOutputTokens *int `pkl:"maxOutputTokens"`
}
