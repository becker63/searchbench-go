// Package policy defines pure policy role vocabulary.
package policy

import "github.com/becker63/searchbench-go/internal/pure/domain"

// IncumbentPolicy is the policy defending its position in a round.
type IncumbentPolicy = domain.SystemRef

// ChallengerPolicy is the policy trying to replace the incumbent in a round.
type ChallengerPolicy = domain.SystemRef

// Pair stores the incumbent/challenger policy roles for one round.
type Pair struct {
	Incumbent  IncumbentPolicy  `json:"incumbent"`
	Challenger ChallengerPolicy `json:"challenger"`
}
