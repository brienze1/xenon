package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const microsecondsPerMillisecond = 1000

type ConsoleResponse struct {
	Error      string  `json:"error,omitempty"`
	StatusText string  `json:"status_text"`
	TimeMs     float64 `json:"time_ms"`
	SizeBytes  int64   `json:"size_bytes"`
	StatusCode int     `json:"status_code"`
}

func NewConsoleResponse(response entity.Response) ConsoleResponse {
	result := ConsoleResponse{StatusCode: response.StatusCode, StatusText: response.StatusText, SizeBytes: response.SizeBytes, TimeMs: float64(response.Timings.Total.Microseconds()) / microsecondsPerMillisecond}
	if response.Error != nil {
		result.Error = response.Error.Message
	}
	return result
}
