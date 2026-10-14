package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPPage[T any] struct {
	Next  string `json:"next,omitempty"`
	Items []T    `json:"items"`
}

func NewMCPPage[E, T any](page entity.AIPage[E], convert func(E) T) MCPPage[T] {
	items := make([]T, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, convert(item))
	}
	return MCPPage[T]{Next: page.Next, Items: items}
}

func NewMCPPageOf[T any](page entity.AIPage[T]) MCPPage[T] {
	return NewMCPPage(page, func(item T) T { return item })
}

func NewMCPList[T any](items []T) MCPPage[T] {
	return MCPPage[T]{Items: append(make([]T, 0, len(items)), items...)}
}
