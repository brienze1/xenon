package model

import "github.com/brienze1/xenon/internal/domain/entity"

type RunSentRequest struct {
	Method  string     `json:"method"`
	Url     string     `json:"url"`
	Body    string     `json:"body"`
	Headers []KeyValue `json:"headers"`
}

func NewRunSentRequest(request entity.SentRequest) RunSentRequest {
	headers := make([]KeyValue, 0, len(request.Headers))
	for _, header := range request.Headers {
		headers = append(headers, NewKeyValue(header))
	}
	return RunSentRequest{Method: request.Method, Url: request.URL, Body: request.Body, Headers: headers}
}

func (r *RunSentRequest) ToEntity() entity.SentRequest {
	headers := make([]entity.KeyValue, 0, len(r.Headers))
	for _, header := range r.Headers {
		headers = append(headers, header.ToEntity())
	}
	return entity.SentRequest{Method: r.Method, URL: r.Url, Body: r.Body, Headers: headers}
}
