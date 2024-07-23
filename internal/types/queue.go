package types

import (
	"errors"
	"sync"
)

type Queue struct {
	Elements []int
	mutex    sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		Elements: []int{},
	}
}

func (q *Queue) Enqueue(elem int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Elements = append(q.Elements, elem)
}

func (q *Queue) Dequeue() (int, error) {
	var zero int // To return in case of underflow

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

func (q *Queue) GetLength() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.Elements)
}

func (q *Queue) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.Elements) == 0
}

func (q *Queue) Peek() (int, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	var zero int // To return in case of empty queue
	if len(q.Elements) == 0 {
		return zero, errors.New("empty queue")
	}
	return q.Elements[0], nil
}
