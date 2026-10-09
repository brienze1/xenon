package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const ResponseBodyRoute = "/xenon/response"

type SendResult struct {
	Error       *SendResultError  `json:"error,omitempty"`
	InlineBody  *string           `json:"inline_body,omitempty"`
	TabId       string            `json:"tab_id"`
	StatusText  string            `json:"status_text"`
	Proto       string            `json:"proto"`
	ContentType string            `json:"content_type"`
	BodyUrl     string            `json:"body_url"`
	Headers     []ResponseHeader  `json:"headers"`
	Timings     SendResultTimings `json:"timings"`
	StatusCode  int               `json:"status_code"`
	SizeBytes   int64             `json:"size_bytes"`
	Ok          bool              `json:"ok"`
	IsBinary    bool              `json:"is_binary"`
}

func NewSendResult(tabId string, response entity.Response) SendResult {
	result := SendResult{
		TabId:       tabId,
		Ok:          response.IsOk(),
		StatusCode:  response.StatusCode,
		StatusText:  response.StatusText,
		Proto:       response.Proto,
		Headers:     make([]ResponseHeader, 0, len(response.Headers)),
		SizeBytes:   response.SizeBytes,
		ContentType: response.ContentType,
		IsBinary:    response.Binary,
		BodyUrl:     ResponseBodyRoute + "/" + tabId,
		Timings:     NewSendResultTimings(response.Timings),
	}
	for _, header := range response.Headers {
		result.Headers = append(result.Headers, NewResponseHeader(header))
	}
	if response.HasInlineBody() {
		body := string(response.Body)
		result.InlineBody = &body
	}
	if !response.IsOk() {
		sendError := NewSendResultError(*response.Error)
		result.Error = &sendError
	}
	return result
}

func (s SendResult) LogValue() any {
	return map[string]any{"tab_id": s.TabId, "ok": s.Ok, "status_code": s.StatusCode, "size_bytes": s.SizeBytes}
}
