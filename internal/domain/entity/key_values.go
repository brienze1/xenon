package entity

type KeyValues []KeyValue

func mapKeyValues(rows []KeyValue, transform func(string) string) []KeyValue {
	mapped := make([]KeyValue, 0, len(rows))
	for _, row := range rows {
		mapped = append(mapped, row.MapText(transform))
	}
	return mapped
}

func (k KeyValues) LogValue() any {
	values := make([]any, 0, len(k))
	for _, row := range k {
		values = append(values, row.LogValue())
	}
	return values
}
