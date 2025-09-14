package concurrent

import (
	"container/heap"
	"sync"
)

// PriorityQueue implements a priority queue for tasks
type PriorityQueue struct {
	items []Task
	mutex sync.RWMutex
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		items: make([]Task, 0),
	}
	heap.Init(pq)
	return pq
}

// Len returns the number of items in the queue
func (pq *PriorityQueue) Len() int {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	return len(pq.items)
}

// Less compares two items for priority (lower priority number = higher priority)
func (pq *PriorityQueue) Less(i, j int) bool {
	return pq.items[i].GetPriority() < pq.items[j].GetPriority()
}

// Swap swaps two items in the queue
func (pq *PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push adds an item to the queue
func (pq *PriorityQueue) Push(x interface{}) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	
	item := x.(Task)
	pq.items = append(pq.items, item)
}

// Pop removes and returns the highest priority item
func (pq *PriorityQueue) Pop() interface{} {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	
	old := pq.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	pq.items = old[0 : n-1]
	return item
}

// Peek returns the highest priority item without removing it
func (pq *PriorityQueue) Peek() Task {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	
	if len(pq.items) == 0 {
		return nil
	}
	return pq.items[0]
}

// IsEmpty returns true if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()
	return len(pq.items) == 0
}