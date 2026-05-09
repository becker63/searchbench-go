package domain

import "iter"

// Role names the two sides of a Searchbench comparison.
type Role string

const (
	RoleIncumbent  Role = "incumbent"
	RoleChallenger Role = "challenger"
)

// Pair stores the incumbent and challenger versions of the same kind of value.
//
// Searchbench is fundamentally comparative, so this type should appear often:
// Pair[SystemSpec], Pair[[]ScoredRun], Pair[ScoreSummary], etc.
type Pair[T any] struct {
	Incumbent  T `json:"incumbent"`
	Challenger T `json:"challenger"`
}

// NewPair constructs the incumbent/challenger pair for a value type.
func NewPair[T any](incumbent T, challenger T) Pair[T] {
	return Pair[T]{
		Incumbent:  incumbent,
		Challenger: challenger,
	}
}

// MapPair maps both sides of the pair while preserving incumbent/challenger
// roles.
func MapPair[A, B any](p Pair[A], f func(Role, A) B) Pair[B] {
	return Pair[B]{
		Incumbent:  f(RoleIncumbent, p.Incumbent),
		Challenger: f(RoleChallenger, p.Challenger),
	}
}

// All iterates incumbent first and then challenger together with their roles.
func (p Pair[T]) All() iter.Seq2[Role, T] {
	return func(yield func(Role, T) bool) {
		if !yield(RoleIncumbent, p.Incumbent) {
			return
		}
		yield(RoleChallenger, p.Challenger)
	}
}

// ByRole returns the value for the requested comparison role.
func (p Pair[T]) ByRole(role Role) T {
	switch role {
	case RoleIncumbent:
		return p.Incumbent
	case RoleChallenger:
		return p.Challenger
	default:
		var zero T
		return zero
	}
}
