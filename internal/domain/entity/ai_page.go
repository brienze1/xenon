package entity

const AIPageSize = 100

type AIPage[T any] struct {
	Next  string
	Items []T
}

// NewAIPage returns up to 100 items after the cursor; false means an unknown cursor.
func NewAIPage[T any](items []T, cursor string) (AIPage[T], bool) {
	offset, valid := ParseAICursor(cursor)
	if !valid {
		return AIPage[T]{}, false
	}
	start := min(offset, len(items))
	end := min(start+AIPageSize, len(items))
	page := AIPage[T]{Items: append(make([]T, 0, end-start), items[start:end]...)}
	if end < len(items) {
		page.Next = encodeAICursor(end)
	}
	return page, true
}
