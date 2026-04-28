package domain

import "iter"

// Role names the two sides of a Searchbench comparison.
type Role string

const (
	RoleBaseline  Role = "baseline"
	RoleCandidate Role = "candidate"
)

// Pair stores the baseline and candidate versions of the same kind of value.
//
// Searchbench is fundamentally comparative, so this type should appear often:
// Pair[SystemSpec], Pair[[]ScoredRun], Pair[ScoreSummary], etc.
type Pair[T any] struct {
	Baseline  T `json:"baseline"`
	Candidate T `json:"candidate"`
}

// NewPair constructs the baseline/candidate pair for a value type.
func NewPair[T any](baseline T, candidate T) Pair[T] {
	return Pair[T]{
		Baseline:  baseline,
		Candidate: candidate,
	}
}

// MapPair maps both sides of the pair while preserving baseline/candidate
// roles.
func MapPair[A, B any](p Pair[A], f func(Role, A) B) Pair[B] {
	return Pair[B]{
		Baseline:  f(RoleBaseline, p.Baseline),
		Candidate: f(RoleCandidate, p.Candidate),
	}
}

// All iterates baseline first and then candidate together with their roles.
func (p Pair[T]) All() iter.Seq2[Role, T] {
	return func(yield func(Role, T) bool) {
		if !yield(RoleBaseline, p.Baseline) {
			return
		}
		yield(RoleCandidate, p.Candidate)
	}
}

// ByRole returns the value for the requested comparison role.
func (p Pair[T]) ByRole(role Role) T {
	switch role {
	case RoleBaseline:
		return p.Baseline
	case RoleCandidate:
		return p.Candidate
	default:
		var zero T
		return zero
	}
}
