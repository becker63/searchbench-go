// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

import "github.com/becker63/searchbench-go/internal/adapters/config/pkl/generated/backend"

// One executable policy: identity, engine backend, prompts, and runtime limits.
type System struct {
	// Stable slug for this system in runs and reports.
	Id string `pkl:"id"`

	// Display name; defaults to `id` when omitted in instances (see default below).
	Name string `pkl:"name"`

	// Which runtime backend executes this system’s policy implementation.
	Backend backend.Backend `pkl:"backend"`

	// Prompt bundle reference (name + optional version) resolved by the harness.
	PromptBundle PromptBundle `pkl:"promptBundle"`

	// Per-run execution bounds (steps, wall clock).
	Runtime Runtime `pkl:"runtime"`
}
