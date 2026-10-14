package dto

import (
	"math"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type MCPSendResponse struct {
	Error       *SendResultError  `json:"error,omitempty"`
	StatusText  string            `json:"status_text,omitempty"`
	Body        string            `json:"body,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	Headers     [][]string        `json:"headers,omitempty"`
	Tests       []MCPSendTestLine `json:"tests,omitempty"`
	Status      int               `json:"status,omitempty"`
	Ms          int64             `json:"ms"`
	Size        int64             `json:"size,omitempty"`
	Truncated   bool              `json:"truncated,omitempty"`
	Binary      bool              `json:"binary,omitempty"`
}

func NewMCPSendResponse(result SendResult) MCPSendResponse {
	response := MCPSendResponse{
		Status:      result.StatusCode,
		StatusText:  result.StatusText,
		Ms:          int64(math.Round(result.Timings.TotalMs)),
		ContentType: result.ContentType,
		Error:       result.Error,
		Size:        result.SizeBytes,
		Binary:      entity.IsAIBinaryBody(result.IsBinary, result.ContentType),
		Headers:     make([][]string, 0, len(result.Headers)),
		Tests:       make([]MCPSendTestLine, 0, len(result.Tests)),
	}
	for _, header := range result.Headers {
		response.Headers = append(response.Headers, []string{header.Key, header.Value})
	}
	for _, test := range result.Tests {
		response.Tests = append(response.Tests, NewMCPSendTestLine(test))
	}
	return response
}

func (m MCPSendResponse) WithBody(body entity.AIInlineBody) MCPSendResponse {
	m.Body, m.Truncated, m.Size = body.Text, body.Truncated, int64(body.Size)
	return m
}
