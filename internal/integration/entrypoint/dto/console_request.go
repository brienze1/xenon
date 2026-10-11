package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ConsoleRequest struct {
	Body    *string          `json:"body,omitempty"`
	Method  string           `json:"method"`
	Url     string           `json:"url"`
	Headers []ResponseHeader `json:"headers,omitempty"`
	Nested  bool             `json:"nested"`
}

func NewConsoleRequest(request entity.EffectiveRequest, nested bool) ConsoleRequest {
	return ConsoleRequest{Method: request.Method, Url: request.URL, Nested: nested}
}

func NewConsoleRequestDetail(request entity.EffectiveRequest, nested bool) ConsoleRequest {
	result := NewConsoleRequest(request, nested)
	result.Headers = make([]ResponseHeader, 0, len(request.SendHeaders()))
	for _, header := range request.SendHeaders() {
		result.Headers = append(result.Headers, NewResponseHeader(header))
	}
	if body := request.Body.PreviewText(); body != "" {
		result.Body = &body
	}
	return result
}
