package entity

type AutoHeaders []AutoHeader

func (a AutoHeaders) LogValue() any {
	values := make([]any, 0, len(a))
	for _, header := range a {
		values = append(values, header.LogValue())
	}
	return values
}
