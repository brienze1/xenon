package dto

func NewUnsavedRequest(url string) Request {
	return Request{Url: url}
}

func NewCreatedRequest(created TreeNode) Request {
	return Request{Id: created.Id, Name: created.Name}
}

// NewEditedRequest drops the stored scripts so only scripts passed in the edit are saved.
func NewEditedRequest(request Request, edit *MCPTreeEdit) Request {
	return Request{
		Auth:          overridden(edit.Auth, request.Auth),
		Settings:      request.Settings,
		Id:            request.Id,
		Name:          overridden(edit.Name, request.Name),
		Description:   overridden(edit.Description, request.Description),
		Method:        overridden(edit.Method, request.Method),
		Url:           overridden(edit.Url, request.Url),
		Revision:      request.Revision,
		Body:          overridden(edit.Body, request.Body),
		Scripts:       edit.Scripts,
		AutoHeaders:   request.AutoHeaders,
		Params:        overridden(edit.Params, request.Params),
		PathVariables: request.PathVariables,
		Headers:       overridden(edit.Headers, request.Headers),
	}
}
