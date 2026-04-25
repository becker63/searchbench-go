package report

import (
	"errors"
	"fmt"
	"iter"

	"github.com/becker63/searchbench-go/internal/domain"
)

type TaskRunSet[T any] struct {
	order  []domain.TaskID
	byTask map[domain.TaskID]T
}

func NewTaskRunSet[T any](items map[domain.TaskID]T, order []domain.TaskID) (TaskRunSet[T], error) {
	if len(order) == 0 {
		return TaskRunSet[T]{}, errors.New("task order must be non-empty")
	}

	ordered := append([]domain.TaskID(nil), order...)
	byTask := make(map[domain.TaskID]T, len(items))
	for id, item := range items {
		byTask[id] = item
	}

	seen := make(map[domain.TaskID]struct{}, len(ordered))
	for _, id := range ordered {
		if _, ok := seen[id]; ok {
			return TaskRunSet[T]{}, fmt.Errorf("duplicate task id in order: %s", id)
		}
		seen[id] = struct{}{}
		if _, ok := byTask[id]; !ok {
			return TaskRunSet[T]{}, fmt.Errorf("missing item for task id: %s", id)
		}
	}

	return TaskRunSet[T]{
		order:  ordered,
		byTask: byTask,
	}, nil
}

func (s TaskRunSet[T]) Get(id domain.TaskID) (T, bool) {
	value, ok := s.byTask[id]
	return value, ok
}

func (s TaskRunSet[T]) Order() []domain.TaskID {
	return append([]domain.TaskID(nil), s.order...)
}

func (s TaskRunSet[T]) Values() []T {
	values := make([]T, 0, len(s.order))
	for _, id := range s.order {
		values = append(values, s.byTask[id])
	}
	return values
}

func (s TaskRunSet[T]) Items() iter.Seq2[domain.TaskID, T] {
	return func(yield func(domain.TaskID, T) bool) {
		for _, id := range s.order {
			value := s.byTask[id]
			if !yield(id, value) {
				return
			}
		}
	}
}

func (s TaskRunSet[T]) Len() int {
	return len(s.order)
}
