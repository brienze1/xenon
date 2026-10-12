package entity

import (
	"slices"
	"strings"
)

type ApiOperation struct {
	RequestID   string
	Method      string
	Path        string
	Server      string
	Summary     string
	Description string
	Tags        []string
	PathParams  []ApiParameter
	QueryParams []ApiParameter
	Headers     []ApiParameter
	Responses   []ApiResponse
	Security    []string
	Body        ApiBody
	Variant     bool
}

func NewApiOperation(documented DocumentedRequest, path string, security []string) ApiOperation {
	request := documented.Request
	operationPath, pathParams := DocumentedPath(path, request.PathVariables)
	operation := ApiOperation{
		RequestID:   request.ID,
		Method:      request.MethodOrDefault(),
		Path:        operationPath,
		Summary:     request.Name,
		Description: request.Description,
		Body:        NewApiBody(request.Body),
		Tags:        []string{},
		PathParams:  pathParams,
		QueryParams: documentedRows(request.Params, KeyValue.IsActive),
		Headers:     documentedRows(request.Headers, IsDocumentedHeader),
		Responses:   NewApiResponses(request.Examples),
		Security:    security,
	}
	if tag := documented.Tag(); tag != "" {
		operation.Tags = []string{tag}
	}
	return operation
}

func (a ApiOperation) Tag() string {
	if len(a.Tags) == 0 {
		return ""
	}
	return a.Tags[0]
}

func (a ApiOperation) Name() string {
	if name := NormalizeItemName(a.Summary); name != "" {
		return name
	}
	return a.Method + " " + a.Path
}

// ApplyTo writes what changed between the baseline and this operation onto the request; false when nothing did.
func (a ApiOperation) ApplyTo(request Request, baseline ApiOperation, server, baselineServer string, newID func() string) (Request, bool) {
	changed := false
	if a.Summary != baseline.Summary && NormalizeItemName(a.Summary) != "" {
		request.Name, changed = NormalizeItemName(a.Summary), true
	}
	if a.Description != baseline.Description {
		request.Description, changed = a.Description, true
	}
	if a.Method != baseline.Method {
		request.Method, changed = a.Method, true
	}
	if server+a.Path != baselineServer+baseline.Path || !slices.Equal(a.PathParams, baseline.PathParams) || !slices.Equal(a.QueryParams, baseline.QueryParams) {
		request, changed = a.withUrl(request, server), true
	}
	if !slices.Equal(a.Headers, baseline.Headers) {
		request.Headers, changed = append(undocumentedRows(request.Headers, IsDocumentedHeader), parameterRows(a.Headers)...), true
	}
	if !a.Body.Equal(baseline.Body) {
		request.Body, changed = a.Body.AppliedTo(request.Body), true
		request.Headers = withContentType(request.Headers, a.Body, baseline.Body)
	}
	if !slices.EqualFunc(a.Responses, baseline.Responses, ApiResponse.Equal) {
		request.Examples, changed = a.examplesApplied(request, baseline, newID), true
	}
	return request, changed
}

func (a ApiOperation) withUrl(request Request, server string) Request {
	base := strings.TrimSuffix(server, "/") + RequestPath(a.Path, a.PathParams)
	request.URL = ParseRequestUrl(base).WithQuery(parameterRows(a.QueryParams)).String()
	request.Params = append(parameterRows(a.QueryParams), undocumentedRows(request.Params, KeyValue.IsActive)...)
	variables := make([]PathVariable, 0, len(a.PathParams))
	for _, parameter := range a.PathParams {
		if !parameter.Variable {
			variables = append(variables, parameter.PathVariable())
		}
	}
	request.PathVariables = variables
	return request
}

func (a ApiOperation) examplesApplied(request Request, baseline ApiOperation, newID func() string) []SavedExample {
	examples := make([]SavedExample, 0)
	for _, response := range a.Responses {
		for _, documented := range response.Examples {
			saved, exists := request.Example(documented.ID)
			if !exists {
				examples = append(examples, documented.NewSavedExample(newID(), response.Code, request))
				continue
			}
			previous, previousCode := baseline.example(documented.ID)
			if previous == documented && previousCode == response.Code {
				examples = append(examples, saved)
				continue
			}
			examples = append(examples, documented.AppliedTo(saved, response.Code, previous))
		}
	}
	return examples
}

func (a ApiOperation) example(id string) (ApiExample, int) {
	for _, response := range a.Responses {
		for _, example := range response.Examples {
			if example.ID == id {
				return example, response.Code
			}
		}
	}
	return ApiExample{}, 0
}

func withContentType(headers []KeyValue, body, baseline ApiBody) []KeyValue {
	if body.MediaType == baseline.MediaType || !body.HasContent() || body.IsForm() {
		return headers
	}
	updated := slices.Clone(headers)
	for index, header := range updated {
		if header.IsActive() && header.HasKey("Content-Type") {
			updated[index].Value = body.MediaType
		}
	}
	return updated
}
