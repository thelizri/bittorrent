package types

import (
	"errors"
	"sync"
)

type Queue[T any] struct {
	Elements []T
	mutex    sync.RWMutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		Elements: []T{},
	}
}

func (q *Queue[T]) Enqueue(elem T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Elements = append(q.Elements, elem)
}

func (q *Queue[T]) Dequeue() (T, error) {
	var zero T // To return in case of underflow

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.Elements) == 0 {
		return zero, errors.New("queue is empty")
	}

	element := q.Elements[0]
	if len(q.Elements) == 1 {
		q.Elements = nil
	} else {
		q.Elements = q.Elements[1:]
	}
	return element, nil
}

func (q *Queue[T]) GetLength() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.Elements)
}

func (q *Queue[T]) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.Elements) == 0
}

func (q *Queue[T]) Peek() (T, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	var zero T // To return in case of empty queue
	if len(q.Elements) == 0 {
		return zero, errors.New("empty queue")
	}
	return q.Elements[0], nil
}
