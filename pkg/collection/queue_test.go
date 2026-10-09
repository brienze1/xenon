package collection

import "testing"

func TestQueueIsFifo(t *testing.T) {
	queue := NewQueue[int]()
	queue.Push(1)
	queue.Push(2)
	for _, want := range []int{1, 2} {
		got, ok := queue.Pop()
		if !ok || got != want {
			t.Fatalf("Pop() = %d, %v; want %d, true", got, ok, want)
		}
	}
	if got, ok := queue.Pop(); ok || got != 0 {
		t.Errorf("Pop() on empty = %d, %v", got, ok)
	}
}
