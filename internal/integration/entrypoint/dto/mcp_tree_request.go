package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPTreeRequest struct {
	Scripts     *Scripts       `json:"scripts,omitempty"`
	Auth        Auth           `json:"auth"`
	Body        Body           `json:"body"`
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Method      string         `json:"method"`
	Url         string         `json:"url"`
	Description string         `json:"description,omitempty"`
	Path        string         `json:"path"`
	Params      []KeyValue     `json:"params"`
	Headers     []KeyValue     `json:"headers"`
	Aliases     []MCPTreeAlias `json:"aliases"`
}

func NewMCPTreeRequest(request Request, path string, environments []entity.Environment) MCPTreeRequest {
	return MCPTreeRequest{
		Id:          request.Id,
		Name:        request.Name,
		Method:      request.Method,
		Url:         request.Url,
		Params:      request.Params,
		Headers:     request.Headers,
		Body:        request.Body,
		Auth:        request.Auth,
		Scripts:     request.Scripts,
		Description: request.Description,
		Path:        path,
		Aliases:     NewMCPTreeAliases(request.Aliases, environments),
	}
}
