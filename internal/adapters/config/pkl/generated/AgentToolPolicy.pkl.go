// Code generated from Pkl module `searchbench.config.SearchBenchRound`. DO NOT EDIT.
package generated

// Allow/deny tool policy for an agent manifest surface.
//
// Structural validation is done in config load; **evaluator** tool names are also checked at round resolution
// against the runtime tool registry (`Available` vs explicit allow vs defaults).
type AgentToolPolicy struct {
	// Explicit allow list; when empty, the runtime uses **default allowed** tools minus **deny** (see round resolution docs).
	Allow []string `pkl:"allow"`

	// Names subtracted from the active candidate set (must not overlap explicit allow entries; validated in Go).
	Deny []string `pkl:"deny"`
}
