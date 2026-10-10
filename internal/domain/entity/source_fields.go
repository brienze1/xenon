package entity

type SourceFields map[string]any

func (s SourceFields) IsEmpty() bool {
	return len(s) == 0
}
