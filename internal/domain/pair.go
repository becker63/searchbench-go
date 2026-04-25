package domain

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

func NewPair[T any](baseline T, candidate T) Pair[T] {
	return Pair[T]{
		Baseline:  baseline,
		Candidate: candidate,
	}
}

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
