package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunSentRequest struct {
	Method  string           `json:"method"`
	Url     string           `json:"url"`
	Body    string           `json:"body"`
	Headers []ResponseHeader `json:"headers"`
}

func NewRunSentRequest(request entity.SentRequest) RunSentRequest {
	headers := make([]ResponseHeader, 0, len(request.Headers))
	for _, header := range request.Headers {
		headers = append(headers, NewResponseHeader(header))
	}
	return RunSentRequest{Method: request.Method, Url: request.URL, Body: request.Body, Headers: headers}
}
