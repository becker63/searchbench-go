package execution

import "github.com/becker63/searchbench-go/internal/pure/domain"

// Spec is a planned request to run one system on one match.
//
// Conceptually:
//
//	Spec = MatchSpec + SystemSpec
//
// It should be deterministic and serializable. Runtime state belongs in
// PreparedRun or ExecutedRun, not here.
type Spec struct {
	ID     domain.RunID      `json:"id"`
	Match  domain.MatchSpec  `json:"match"`
	System domain.SystemSpec `json:"system"`
}

// NewSpec constructs a planned run request from one match and one executable
// system.
func NewSpec(id domain.RunID, match domain.MatchSpec, system domain.SystemSpec) Spec {
	return Spec{
		ID:     id,
		Match:  match,
		System: system,
	}
}
