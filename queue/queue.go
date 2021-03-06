package queue

import (
	"fmt"
	"sync"
)

// Queue represents a queue
type Queue struct {
	lock sync.RWMutex
	data []interface{}
}

var (
	// ErrorEmpty is returned on illegal operations on an empty queue
	ErrorEmpty = fmt.Errorf("empty queue")
)

// Len returns the number of items in the queue
func (q *Queue) Len() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.data)
}

// Add adds an item at the end of the queue
func (q *Queue) Add(item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.data = append(q.data, item)
}

// Peek returns the first item from the queue without removing it
func (q *Queue) Peek() (interface{}, error) {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.data) == 0 {
		return nil, ErrorEmpty
	}
	return q.data[0], nil
}

// Remove returns the first item from the queue
func (q *Queue) Remove() (interface{}, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.data) == 0 {
		return nil, ErrorEmpty
	}
	item := q.data[0]
	q.data = q.data[1:]
	return item, nil
}
