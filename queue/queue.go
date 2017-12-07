package queue

import "sync"

// Queue contains an ordered list of byte slices
type Queue struct {
	items [][]byte
	sync.Mutex
}

// New creates a new empty Queue
func New() *Queue {
	q := new(Queue)
	q.items = make([][]byte, 0, 1024)
	return q
}

// Push a slice of bytes onto the queue
func (q *Queue) Push(item []byte) {
	q.items = append(q.items, item)
}

// Pop a slice off bytes off the queue
func (q *Queue) Pop() []byte {
	q.Lock()
	item := q.items[0]
	q.items = q.items[1:]
	q.Unlock()
	return item
}

// Length returns the number of byte slices in the queue
func (q *Queue) Length() int {
	return len(q.items)
}
