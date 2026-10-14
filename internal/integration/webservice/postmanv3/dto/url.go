package dto

func urlToPostmanV21(raw string, query, variables KeyValues) any {
	if query == nil && variables == nil {
		return raw
	}
	url := map[string]any{"raw": raw}
	setIfPresent(url, "query", query.toPostmanV21())
	setIfPresent(url, "variable", pathVariablesToPostmanV21(variables))
	return url
}

func pathVariablesToPostmanV21(variables KeyValues) []any {
	rows := make([]any, 0, len(variables))
	for index := range variables {
		row := variables[index].toPostmanV21()
		delete(row, "disabled")
		rows = append(rows, row)
	}
	return rows
}
