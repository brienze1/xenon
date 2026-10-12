package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiOperation struct {
	RequestBody *ApiBody       `json:"request_body"`
	RequestId   string         `json:"request_id"`
	Tag         string         `json:"tag"`
	Method      string         `json:"method"`
	Path        string         `json:"path"`
	Server      string         `json:"server"`
	Summary     string         `json:"summary"`
	Description string         `json:"description"`
	Parameters  []ApiParameter `json:"parameters"`
	Responses   []ApiResponse  `json:"responses"`
	Security    []string       `json:"security"`
	Variant     bool           `json:"variant"`
}

func NewApiOperation(operation entity.ApiOperation, server string) ApiOperation {
	documented := ApiOperation{
		RequestId:   operation.RequestID,
		Tag:         operation.Tag(),
		Method:      operation.Method,
		Path:        operation.Path,
		Server:      server,
		Summary:     operation.Summary,
		Description: operation.Description,
		Parameters:  make([]ApiParameter, 0, len(operation.PathParams)+len(operation.QueryParams)+len(operation.Headers)),
		Responses:   make([]ApiResponse, 0, len(operation.Responses)),
		Security:    append([]string{}, operation.Security...),
		Variant:     operation.Variant,
	}
	for _, parameter := range operation.PathParams {
		documented.Parameters = append(documented.Parameters, NewApiParameter(parameter, "path"))
	}
	for _, parameter := range operation.QueryParams {
		documented.Parameters = append(documented.Parameters, NewApiParameter(parameter, "query"))
	}
	for _, parameter := range operation.Headers {
		documented.Parameters = append(documented.Parameters, NewApiParameter(parameter, "header"))
	}
	if operation.Body.HasContent() {
		body := NewApiBody(operation.Body)
		documented.RequestBody = &body
	}
	for _, response := range operation.Responses {
		documented.Responses = append(documented.Responses, NewApiResponse(response))
	}
	return documented
}
