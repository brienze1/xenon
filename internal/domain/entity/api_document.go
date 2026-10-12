package entity

import "slices"

type ApiDocument struct {
	Title           string
	Description     string
	Version         string
	Servers         []string
	Tags            []ApiTag
	Operations      []ApiOperation
	SecuritySchemes []ApiSecurityScheme
}

// ServerOf is where the operation is served: its own server, else the document's first one.
func (a ApiDocument) ServerOf(operation ApiOperation) string {
	if operation.Server != "" || len(a.Servers) == 0 {
		return operation.Server
	}
	return a.Servers[0]
}

func (a ApiDocument) Operation(requestID string) (ApiOperation, bool) {
	index := slices.IndexFunc(a.Operations, func(operation ApiOperation) bool { return operation.RequestID == requestID })
	if index < 0 {
		return ApiOperation{}, false
	}
	return a.Operations[index], true
}
