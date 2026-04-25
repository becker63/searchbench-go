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

// NewNonEmpty constructs a non-empty collection from a required head element.
func NewNonEmpty[T any](head T, tail ...T) NonEmpty[T] {
	return NonEmpty[T]{
		head:  head,
		tail:  append([]T(nil), tail...),
		valid: true,
	}
}

// NonEmptyFromSlice constructs a NonEmpty value from a slice, rejecting empty
// input.
func NonEmptyFromSlice[T any](items []T) (NonEmpty[T], error) {
	if len(items) == 0 {
		return NonEmpty[T]{}, errEmptyNonEmpty
	}
	return NewNonEmpty(items[0], items[1:]...), nil
}

// Head returns the first element.
func (n NonEmpty[T]) Head() T {
	return n.head
}

// Tail returns the elements after Head.
func (n NonEmpty[T]) Tail() []T {
	if !n.valid {
		return nil
	}
	return append([]T(nil), n.tail...)
}

// All returns the full collection as a slice copy.
func (n NonEmpty[T]) All() []T {
	if !n.valid {
		return nil
	}
	items := make([]T, 1+len(n.tail))
	items[0] = n.head
	copy(items[1:], n.tail)
	return items
}

// Values iterates the collection in stable order.
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

// Len reports the number of elements, or zero for an invalid zero value.
func (n NonEmpty[T]) Len() int {
	if !n.valid {
		return 0
	}
	return 1 + len(n.tail)
}

// Valid reports whether the value was constructed as a meaningful non-empty
// collection.
func (n NonEmpty[T]) Valid() bool {
	return n.valid
}

// Validate rejects the invalid zero value.
func (n NonEmpty[T]) Validate() error {
	if !n.valid {
		return errEmptyNonEmpty
	}
	return nil
}

// MarshalJSON encodes the collection as a JSON array.
func (n NonEmpty[T]) MarshalJSON() ([]byte, error) {
	if err := n.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(n.All())
}

// UnmarshalJSON decodes a JSON array into a validated non-empty collection.
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
