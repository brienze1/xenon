package dto

func NewEditedExampleRequest(request ExampleRequest, edit *MCPTreeEdit) ExampleRequest {
	return ExampleRequest{
		Auth:          overridden(edit.Auth, request.Auth),
		Method:        overridden(edit.Method, request.Method),
		Url:           overridden(edit.Url, request.Url),
		Body:          overridden(edit.Body, request.Body),
		Params:        overridden(edit.Params, request.Params),
		PathVariables: request.PathVariables,
		Headers:       overridden(edit.Headers, request.Headers),
	}
}
