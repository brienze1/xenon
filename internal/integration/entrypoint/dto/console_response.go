package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ConsoleResponse struct {
	Error       *SendResultError `json:"error,omitempty"`
	Body        *string          `json:"body,omitempty"`
	StatusText  string           `json:"status_text"`
	ContentType string           `json:"content_type,omitempty"`
	BodyUrl     string           `json:"body_url,omitempty"`
	Headers     []ResponseHeader `json:"headers,omitempty"`
	TimeMs      float64          `json:"time_ms"`
	SizeBytes   int64            `json:"size_bytes"`
	StatusCode  int              `json:"status_code"`
	IsBinary    bool             `json:"is_binary"`
}

func NewConsoleResponse(response entity.Response) ConsoleResponse {
	result := ConsoleResponse{StatusCode: response.StatusCode, StatusText: response.StatusText, SizeBytes: response.SizeBytes, TimeMs: milliseconds(response.Timings.Total), IsBinary: response.Binary}
	if response.Error != nil {
		sendError := NewSendResultError(*response.Error)
		result.Error = &sendError
	}
	return result
}

// NewConsoleResponseDetail adds headers and the body, inline up to 1 MB of text or else as bodyUrl.
func NewConsoleResponseDetail(response entity.Response, bodyUrl string) ConsoleResponse {
	result := NewConsoleResponse(response)
	result.ContentType = response.ContentType
	result.Headers = make([]ResponseHeader, 0, len(response.Headers))
	for _, header := range response.Headers {
		result.Headers = append(result.Headers, NewResponseHeader(header))
	}
	if response.HasInlineBody() {
		body := string(response.Body)
		result.Body = &body
	} else if response.IsOk() {
		result.BodyUrl = bodyUrl
	}
	return result
}
