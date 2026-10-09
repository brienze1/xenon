package entity

type KeyValues []KeyValue

func (k KeyValues) LogValue() any {
	values := make([]any, 0, len(k))
	for _, row := range k {
		values = append(values, row.LogValue())
	}
	return values
}
