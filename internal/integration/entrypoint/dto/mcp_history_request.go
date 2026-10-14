package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPHistoryRequest struct {
	Auth    *Auth      `json:"auth,omitempty"`
	Body    *Body      `json:"body,omitempty"`
	Name    string     `json:"name,omitempty"`
	Method  string     `json:"method"`
	Url     string     `json:"url"`
	Params  []KeyValue `json:"params,omitempty"`
	Headers []KeyValue `json:"headers,omitempty"`
}

func NewMCPHistoryRequest(sent entity.Request) MCPHistoryRequest {
	request := NewRequest(sent)
	row := MCPHistoryRequest{Name: request.Name, Method: request.Method, Url: request.Url, Params: request.Params, Headers: request.Headers}
	if sent.Auth.IsOwn() {
		row.Auth = &request.Auth
	}
	if sent.Body.HasContent() {
		row.Body = &request.Body
	}
	return row
}
