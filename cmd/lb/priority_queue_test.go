package main

import (
	"errors"
	"github.com/KatePril/architecture-lab-5/cmd/lb/priorityQueue"
	"testing"
)

func TestPriorityQueuePush(t *testing.T) {
	queue := priorityQueue.New()

	tests := []struct {
		name, value string
		priority    int64
		expect      error
	}{
		{"test successful insertion", "server1:8080", 20, nil},
		{"test successful insertion", "server2:8080", 100, nil},
		{"test error insertion", "server1:8080", 20, errors.New("Item duplication: server1:8080")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := queue.Push(tt.value, tt.priority)
			if (result == nil) != (tt.expect == nil) || (result != nil && result.Error() != tt.expect.Error()) {
				t.Errorf("got %v, want %v", result, tt.expect)
			}
		})
	}

}

func TestPriorityQueuePop(t *testing.T) {
	queue := priorityQueue.New()
	t.Run("test pop empty queue", func(t *testing.T) {
		result, err := queue.Pop()
		if err == nil {
			t.Errorf("got %v, want %v", errors.New("queue is empty"), nil)
		}
		if result != "" {
			t.Errorf("got %v, want %v", result, "empty string")
		}
	})

	t.Run("test successful pop", func(t *testing.T) {
		_ = queue.Push("server1:8080", 20)
		result, err := queue.Pop()
		if err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
		if result != "server1:8080" {
			t.Errorf("got %v, want %v", result, "server1:8080")
		}
	})
}

func TestPriorityQueueRemove(t *testing.T) {
	queue := priorityQueue.New()
	_ = queue.Push("server1:8080", 20)

	tests := []struct {
		name, value string
		error       error
	}{
		{"test successful remove", "server1:8080", nil},
		{"test error remove", "server2:8080", errors.New("Item 'server2:8080' does not exist")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queue.Remove(tt.value)
			if (err == nil) != (tt.error == nil) || (err != nil && err.Error() != tt.error.Error()) {
				t.Errorf("got %v, want %v", err, tt.error)
			}
		})
	}
}

func TestPriorityQueueUpdate(t *testing.T) {
	queue := priorityQueue.New()
	_ = queue.Push("server2:8080", 100)
	_ = queue.Push("server1:8080", 20)

	tests := []struct {
		name, value string
		priority    int64
		error       error
	}{
		{"test successful update", "server1:8080", 100, nil},
		{"test successful update", "server2:8080", 100, nil},
		{"test error update", "server3:8080", 20, errors.New("Item 'server3:8080' does not exist")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := queue.Update(tt.value, tt.priority)
			if (err == nil) != (tt.error == nil) || (err != nil && err.Error() != tt.error.Error()) {
				t.Errorf("got %v, want %v", err, tt.error)
			}
		})
	}
}

func TestPriorityQueueBack(t *testing.T) {
	queue := priorityQueue.New()

	t.Run("test error back", func(t *testing.T) {
		result, err := queue.Back()
		if err == nil {
			t.Errorf("got %v, want %v", errors.New("queue is empty"), nil)
		}
		if result != "" {
			t.Errorf("got %v, want %v", result, "empty string")
		}
	})

	t.Run("test successful back", func(t *testing.T) {
		_ = queue.Push("server2:8080", 100)
		_ = queue.Push("server1:8080", 20)
		result, err := queue.Back()

		if err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
		if result != "server2:8080" {
			t.Errorf("got %v, want %v", result, "server2:8080")
		}
	})
}

func TestPriorityQueueFront(t *testing.T) {
	queue := priorityQueue.New()

	t.Run("test error front", func(t *testing.T) {
		result, err := queue.Front()
		if err == nil {
			t.Errorf("got %v, want %v", errors.New("queue is empty"), nil)
		}
		if result != "" {
			t.Errorf("got %v, want %v", result, "empty string")
		}
	})

	t.Run("test successful front", func(t *testing.T) {
		_ = queue.Push("server2:8080", 100)
		_ = queue.Push("server1:8080", 20)
		result, err := queue.Front()

		if err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
		if result != "server1:8080" {
			t.Errorf("got %v, want %v", result, "server1:8080")
		}
	})
}

func TestPriorityQueueExists(t *testing.T) {
	queue := priorityQueue.New()
	_ = queue.Push("server1:8080", 20)

	tests := []struct {
		name, value string
		expected    bool
	}{
		{"test value exists", "server1:8080", true},
		{"test value doesn't exist", "server3:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := queue.Exists(tt.value)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
