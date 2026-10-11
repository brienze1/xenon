package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ConsoleRequest struct {
	Method string `json:"method"`
	Url    string `json:"url"`
	Nested bool   `json:"nested"`
}

func NewConsoleRequest(request entity.EffectiveRequest, nested bool) ConsoleRequest {
	return ConsoleRequest{Method: request.Method, Url: request.URL, Nested: nested}
}
