package priorityQueue

import (
	"container/list"
	"errors"
)

type Pair struct {
	value    string
	priority int64
}

// It's not a universal priority queue specifically written for our case.
// It uses some checks that can significantly affect performance with a large
// number of items. Use it only for Load Balancer
type PriorityQueue struct {
	values *list.List
}

// New initializes new priority queue
func New() *PriorityQueue {
	return &PriorityQueue{list.New().Init()}
}

// Clear clears elements of queue
func (queue *PriorityQueue) Clear() {
	queue.values.Init()
}

// find returns index of the given value
func (queue *PriorityQueue) find(value string) (*list.Element, bool) {
	for iter := queue.values.Front(); iter != nil; iter = iter.Next() {
		pair, _ := iter.Value.(Pair)
		if pair.value == value {
			return iter, true
		}
	}
	return nil, false
}

// Push adds element to queue
func (queue *PriorityQueue) Push(value string, priority int64) error {
	if _, found := queue.find(value); found {
		return errors.New("Item duplication: " + value)
	}
	pair := Pair{value, priority}
	for iter := queue.values.Front(); iter != nil; iter = iter.Next() {
		if priority < iter.Value.(Pair).priority {
			queue.values.InsertBefore(pair, iter)
			return nil
		}
	}
	queue.values.PushBack(Pair{value, priority})
	return nil
}

// Pop removes the last element from queue
func (queue *PriorityQueue) Pop() (string, error) {
	if queue.values.Len() == 0 {
		return "", errors.New("queue is empty")
	}
	element := queue.values.Back()
	queue.values.Remove(element)
	pair, _ := element.Value.(Pair)
	return pair.value, nil
}

// Remove removes element with given value from the queue
func (queue *PriorityQueue) Remove(value string) error {
	element, found := queue.find(value)
	if !found {
		return errors.New("Item '" + value + "' does not exist")
	}
	queue.values.Remove(element)
	return nil
}

// Update modifies order of values according to their priorities
func (queue *PriorityQueue) Update(value string, priority int64) error {
	element, found := queue.find(value)
	if !found {
		return errors.New("Item '" + value + "' does not exist")
	}
	pair, _ := element.Value.(Pair)
	pair.priority += priority
	for iter := queue.values.Front(); iter != nil; iter = iter.Next() {
		if pair.priority < iter.Value.(Pair).priority {
			queue.values.MoveBefore(element, iter)
			return nil
		}
	}
	queue.values.MoveAfter(element, queue.values.Back())
	return nil
}

// Back returns the last element of queue
func (queue *PriorityQueue) Back() (string, error) {
	if queue.values.Len() == 0 {
		return "", errors.New("queue is empty")
	}
	pair, _ := queue.values.Back().Value.(Pair)
	return pair.value, nil
}

// Front returns the first element of queue
func (queue *PriorityQueue) Front() (string, error) {
	if queue.values.Len() == 0 {
		return "", errors.New("queue is empty")
	}
	pair, _ := queue.values.Front().Value.(Pair)
	return pair.value, nil
}

// Exists checks if the element exists in queue
func (queue *PriorityQueue) Exists(value string) bool {
	_, exists := queue.find(value)
	return exists
}
