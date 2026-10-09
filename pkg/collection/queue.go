package collection

type Queue[T any] interface {
	Push(item T)
	Pop() (T, bool)
}

type queue[T any] struct {
	items []T
}

func NewQueue[T any]() Queue[T] {
	return &queue[T]{items: make([]T, 0)}
}

func (q *queue[T]) Push(item T) {
	q.items = append(q.items, item)
}

func (q *queue[T]) Pop() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}
