package dto

import "github.com/brienze1/xenon/internal/domain/entity"

// ScriptResponse is a response as a script sees it (Postman's response JSON shape).
type ScriptResponse struct {
	Cookie       []ScriptSandboxCookie `json:"cookie,omitempty"`
	Status       string                `json:"status"`
	Body         string                `json:"body"`
	Header       []ScriptKeyValue      `json:"header"`
	Code         int                   `json:"code"`
	ResponseTime int64                 `json:"responseTime"`
	ResponseSize int64                 `json:"responseSize,omitempty"`
}

func NewScriptResponse(response entity.Response) ScriptResponse {
	return ScriptResponse{
		Code:         response.StatusCode,
		Status:       response.StatusText,
		Header:       NewScriptKeyValues(response.Headers),
		Body:         string(response.Body),
		ResponseTime: response.Timings.Total.Milliseconds(),
		ResponseSize: response.SizeBytes,
	}
}
