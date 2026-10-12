package entity

import (
	"slices"
	"strings"
)

var undocumentedHeaders = []string{"accept", "authorization", "content-type", "cookie", "proxy-authorization"}

type ApiParameter struct {
	Name        string
	Description string
	Example     string
	Required    bool
	Variable    bool
}

func NewApiParameter(row KeyValue) ApiParameter {
	return ApiParameter{Name: row.Key, Description: row.Description, Example: row.Value}
}

func (a ApiParameter) KeyValue() KeyValue {
	row := NewKeyValue(a.Name, a.Example)
	row.Description = a.Description
	return row
}

func (a ApiParameter) PathVariable() PathVariable {
	return PathVariable{Key: a.Name, Value: a.Example, Description: a.Description}
}

func IsDocumentedHeader(row KeyValue) bool {
	return row.IsActive() && !slices.Contains(undocumentedHeaders, strings.ToLower(strings.TrimSpace(row.Key)))
}

func documentedRows(rows []KeyValue, documented func(KeyValue) bool) []ApiParameter {
	parameters := make([]ApiParameter, 0, len(rows))
	for _, row := range rows {
		if documented(row) {
			parameters = append(parameters, NewApiParameter(row))
		}
	}
	return parameters
}

func undocumentedRows(rows []KeyValue, documented func(KeyValue) bool) []KeyValue {
	kept := make([]KeyValue, 0)
	for _, row := range rows {
		if !documented(row) {
			kept = append(kept, row)
		}
	}
	return kept
}

func parameterRows(parameters []ApiParameter) []KeyValue {
	rows := make([]KeyValue, 0, len(parameters))
	for _, parameter := range parameters {
		rows = append(rows, parameter.KeyValue())
	}
	return rows
}
