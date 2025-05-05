package priorityQueue

import (
	"container/list"
	"errors"
)

type Pair struct {
	value string
	priority int64
}

// It's not a universal priority queue specifically written for our case.
// It uses some checks that can significantly affect performance with a large
// number of items. Use it only for Load Balancer 
type PriorityQueue struct {
	values *list.List
}

func New() *PriorityQueue {
	return &PriorityQueue{ list.New().Init() }
}

func (queue *PriorityQueue) Clear() {
	queue.values.Init()
}

func (queue *PriorityQueue) find(value string) (*list.Element, bool) {
	for iter := queue.values.Front(); iter != nil; iter = iter.Next() {
		pair, _ := iter.Value.(Pair)
		if pair.value == value {
			return iter, true
		}
	}
	return nil, false
}

func (queue *PriorityQueue) Push(value string, priority int64) error {
	if _, finded := queue.find(value); finded {
		return errors.New("Item duplication: " + value)
	}
	pair := Pair{ value, priority }
	for iter := queue.values.Front(); iter != nil; iter = iter.Next() {
		if priority < iter.Value.(Pair).priority {
			queue.values.InsertBefore(pair, iter)
			return nil
		}
	}
	queue.values.PushBack(Pair{ value, priority })
	return nil
}

func (queue *PriorityQueue) Pop() (string, error) {
	if queue.values.Len() == 0 {
		return "", errors.New("queue is empty")
	}
	element := queue.values.Back()
	queue.values.Remove(element)
	pair, _ := element.Value.(Pair)
	return pair.value, nil
}

func (queue *PriorityQueue) Remove(value string) error {
	element, finded := queue.find(value)
	if !finded {
		return errors.New("Item '" + value + "' does not exist")
	}
	queue.values.Remove(element)
	return nil
}

func (queue *PriorityQueue) Update(value string, priority int64) error {
	element, finded := queue.find(value)
	if !finded {
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

func (queue *PriorityQueue) Back() (string, error) {
	if queue.values.Len() == 0 {
		return "", errors.New("queue is empty")
	}
	pair, _ := queue.values.Back().Value.(Pair)
	return pair.value, nil
}

func (queue *PriorityQueue) Front() (string, error) {
	if queue.values.Len() == 0 {
		return "", errors.New("queue is empty")
	}
	pair, _ := queue.values.Front().Value.(Pair)
	return pair.value, nil
}

func (queue* PriorityQueue) Exists(value string) bool {
	_, exists := queue.find(value)
	return exists
}
