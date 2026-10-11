package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const ResponseBodyRoute = "/xenon/response"

type SendResult struct {
	Error        *SendResultError  `json:"error,omitempty"`
	InlineBody   *string           `json:"inline_body,omitempty"`
	TabId        string            `json:"tab_id"`
	SendId       string            `json:"send_id"`
	Outcome      string            `json:"outcome"`
	StatusText   string            `json:"status_text"`
	Proto        string            `json:"proto"`
	ContentType  string            `json:"content_type"`
	BodyUrl      string            `json:"body_url"`
	Tests        []TestResult      `json:"tests"`
	ScriptErrors []ScriptError     `json:"script_errors"`
	Headers      []ResponseHeader  `json:"headers"`
	CookiesSet   []Cookie          `json:"cookies_set"`
	Timings      SendResultTimings `json:"timings"`
	TestSummary  TestSummary       `json:"test_summary"`
	StatusCode   int               `json:"status_code"`
	SizeBytes    int64             `json:"size_bytes"`
	Ok           bool              `json:"ok"`
	IsBinary     bool              `json:"is_binary"`
}

func NewSendResult(tabId string, execution entity.SendExecution) SendResult {
	result := SendResult{
		TabId:        tabId,
		SendId:       execution.ID,
		Outcome:      execution.Outcome.String(),
		Headers:      []ResponseHeader{},
		CookiesSet:   []Cookie{},
		Tests:        make([]TestResult, 0, len(execution.Tests)),
		TestSummary:  NewTestSummary(execution.TestSummary()),
		ScriptErrors: []ScriptError{},
	}
	for _, test := range execution.Tests {
		result.Tests = append(result.Tests, NewTestResult(test))
	}
	for _, scriptError := range execution.ScriptErrors() {
		result.ScriptErrors = append(result.ScriptErrors, NewScriptError(scriptError))
	}
	if execution.Response != nil {
		result = result.withResponse(*execution.Response)
	}
	return result
}

// NewLastSendResult rebuilds the result of a tab's last send from its kept response.
func NewLastSendResult(tabId string, response entity.Response) SendResult {
	execution := entity.NewSendExecution("")
	execution.RecordResponse(entity.EffectiveRequest{}, response)
	return NewSendResult(tabId, execution)
}

func (s SendResult) withResponse(response entity.Response) SendResult {
	result := SendResult{
		TabId:        s.TabId,
		SendId:       s.SendId,
		Outcome:      s.Outcome,
		Tests:        s.Tests,
		TestSummary:  s.TestSummary,
		ScriptErrors: s.ScriptErrors,
		Ok:           response.IsOk(),
		StatusCode:   response.StatusCode,
		StatusText:   response.StatusText,
		Proto:        response.Proto,
		Headers:      make([]ResponseHeader, 0, len(response.Headers)),
		CookiesSet:   NewCookies(response.Cookies),
		SizeBytes:    response.SizeBytes,
		ContentType:  response.ContentType,
		IsBinary:     response.Binary,
		BodyUrl:      ResponseBodyRoute + "/" + s.TabId,
		Timings:      NewSendResultTimings(response.Timings),
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
	return map[string]any{"tab_id": s.TabId, "send_id": s.SendId, "outcome": s.Outcome, "ok": s.Ok, "status_code": s.StatusCode, "size_bytes": s.SizeBytes}
}
