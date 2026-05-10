package report

import (
	"errors"
	"fmt"
	"iter"

	"github.com/becker63/searchbench-go/internal/pure/domain"
)

type MatchExecutionSet[T any] struct {
	order   []domain.MatchID
	byMatch map[domain.MatchID]T
}

// NewMatchExecutionSet constructs an ordered, match-aligned run set.
//
// It preserves the caller's match order while also enabling lookup by MatchID so
// future report logic can compare incumbent and challenger executions deterministically.
func NewMatchExecutionSet[T any](items map[domain.MatchID]T, order []domain.MatchID) (MatchExecutionSet[T], error) {
	if len(order) == 0 {
		return MatchExecutionSet[T]{}, errors.New("match order must be non-empty")
	}

	ordered := append([]domain.MatchID(nil), order...)
	byMatch := make(map[domain.MatchID]T, len(items))
	for id, item := range items {
		byMatch[id] = item
	}

	seen := make(map[domain.MatchID]struct{}, len(ordered))
	for _, id := range ordered {
		if _, ok := seen[id]; ok {
			return MatchExecutionSet[T]{}, fmt.Errorf("duplicate match id in order: %s", id)
		}
		seen[id] = struct{}{}
		if _, ok := byMatch[id]; !ok {
			return MatchExecutionSet[T]{}, fmt.Errorf("missing item for match id: %s", id)
		}
	}

	return MatchExecutionSet[T]{
		order:   ordered,
		byMatch: byMatch,
	}, nil
}

// Get returns the item for one match ID.
func (s MatchExecutionSet[T]) Get(id domain.MatchID) (T, bool) {
	value, ok := s.byMatch[id]
	return value, ok
}

// Order returns the preserved match ordering.
func (s MatchExecutionSet[T]) Order() []domain.MatchID {
	return append([]domain.MatchID(nil), s.order...)
}

// Values returns the ordered values without their match IDs.
func (s MatchExecutionSet[T]) Values() []T {
	values := make([]T, 0, len(s.order))
	for _, id := range s.order {
		values = append(values, s.byMatch[id])
	}
	return values
}

// Items iterates match/value pairs in preserved match order.
func (s MatchExecutionSet[T]) Items() iter.Seq2[domain.MatchID, T] {
	return func(yield func(domain.MatchID, T) bool) {
		for _, id := range s.order {
			value := s.byMatch[id]
			if !yield(id, value) {
				return
			}
		}
	}
}

// Len reports the number of ordered match entries.
func (s MatchExecutionSet[T]) Len() int {
	return len(s.order)
}
