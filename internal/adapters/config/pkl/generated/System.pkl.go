// Code generated from Pkl module `searchbench.config.Experiment`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/backend"

type System struct {
	Id string `pkl:"id"`

	Name string `pkl:"name"`

	Backend backend.Backend `pkl:"backend"`

	PromptBundle PromptBundle `pkl:"promptBundle"`

	Runtime Runtime `pkl:"runtime"`

	Policy *Policy `pkl:"policy"`
}
