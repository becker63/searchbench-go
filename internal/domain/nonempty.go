package domain

import (
	"encoding/json"
	"errors"
	"iter"
)

var errEmptyNonEmpty = errors.New("non-empty slice is required")

// NonEmpty represents a slice with at least one element.
//
// The zero value can still exist because of Go, so callers should prefer the
// constructors for meaningful values.
type NonEmpty[T any] struct {
	head  T
	tail  []T
	valid bool
}

func NewNonEmpty[T any](head T, tail ...T) NonEmpty[T] {
	return NonEmpty[T]{
		head:  head,
		tail:  append([]T(nil), tail...),
		valid: true,
	}
}

func NonEmptyFromSlice[T any](items []T) (NonEmpty[T], error) {
	if len(items) == 0 {
		return NonEmpty[T]{}, errEmptyNonEmpty
	}
	return NewNonEmpty(items[0], items[1:]...), nil
}

func (n NonEmpty[T]) Head() T {
	return n.head
}

func (n NonEmpty[T]) Tail() []T {
	if !n.valid {
		return nil
	}
	return append([]T(nil), n.tail...)
}

func (n NonEmpty[T]) All() []T {
	if !n.valid {
		return nil
	}
	items := make([]T, 1+len(n.tail))
	items[0] = n.head
	copy(items[1:], n.tail)
	return items
}

func (n NonEmpty[T]) Values() iter.Seq[T] {
	return func(yield func(T) bool) {
		if !n.valid {
			return
		}
		if !yield(n.head) {
			return
		}
		for _, item := range n.tail {
			if !yield(item) {
				return
			}
		}
	}
}

func (n NonEmpty[T]) Len() int {
	if !n.valid {
		return 0
	}
	return 1 + len(n.tail)
}

func (n NonEmpty[T]) Valid() bool {
	return n.valid
}

func (n NonEmpty[T]) Validate() error {
	if !n.valid {
		return errEmptyNonEmpty
	}
	return nil
}

func (n NonEmpty[T]) MarshalJSON() ([]byte, error) {
	if err := n.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(n.All())
}

func (n *NonEmpty[T]) UnmarshalJSON(data []byte) error {
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	value, err := NonEmptyFromSlice(items)
	if err != nil {
		return err
	}
	*n = value
	return nil
}
