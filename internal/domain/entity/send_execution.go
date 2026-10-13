package entity

import (
	"maps"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
)

type SendExecution struct {
	Vault       map[string]string
	SentRequest *EffectiveRequest
	Response    *Response
	Outcome     enums.SendOutcome
	ID          string
	PreRuns     []ScriptRun
	PostRuns    []ScriptRun
	Tests       []TestResult
	Nested      []NestedRequest
	Scopes      ScriptScopes
	Request     Request
}

func NewSendExecution(id string) SendExecution {
	return SendExecution{
		ID:       id,
		Vault:    map[string]string{},
		PreRuns:  []ScriptRun{},
		PostRuns: []ScriptRun{},
		Tests:    []TestResult{},
		Nested:   []NestedRequest{},
	}
}

func (s *SendExecution) Start(request Request, scopes ScriptScopes, vault map[string]string) {
	s.Request = request
	s.Scopes = scopes
	s.Vault = maps.Clone(vault)
}

// ScriptContext is what the next script sees: the request so far, the response, and the scopes and vault as earlier scripts left them.
func (s *SendExecution) ScriptContext(base ScriptContext) ScriptContext {
	base.Request = s.Request
	base.Response = s.Response
	base.Scopes = s.Scopes
	base.Vault = maps.Clone(s.Vault)
	return base
}

func (s *SendExecution) Variables(context VariableContext) VariableContext {
	return context.WithScriptScopes(s.Scopes)
}

func (s *SendExecution) AddRun(run ScriptRun) {
	if run.Script.IsPreRequest() {
		s.PreRuns = append(s.PreRuns, run)
		s.Request = run.Request
	} else {
		s.PostRuns = append(s.PostRuns, run)
	}
	s.Scopes = run.Scopes
	s.Vault = maps.Clone(run.Vault)
	s.Nested = append(s.Nested, run.Nested...)
	s.AddTests(run.Tests)
	if run.HasFailed() && run.StopsSend() {
		s.Outcome = sendoutcome.StoppedByScriptError
	}
	if run.HasSkippedRequest() && run.StopsSend() {
		s.Outcome = sendoutcome.SkippedByScript
	}
}

func (s *SendExecution) RunsPostScripts() bool {
	return s.Outcome == sendoutcome.Sent
}

func (s *SendExecution) ShouldSend() bool {
	return s.Outcome != sendoutcome.StoppedByScriptError && s.Outcome != sendoutcome.SkippedByScript
}

func (s *SendExecution) RecordResponse(sent EffectiveRequest, response Response) {
	s.SentRequest = &sent
	s.Response = &response
	s.Outcome = sendoutcome.Sent
	if response.Error != nil {
		s.Outcome = sendoutcome.Error
	}
	if response.Error != nil && response.Error.Kind == senderrorkind.Cancelled {
		s.Outcome = sendoutcome.Cancelled
	}
}

func (s *SendExecution) ScriptCookies() []ScriptCookie {
	if s.Response == nil || s.SentRequest == nil {
		return []ScriptCookie{}
	}
	return NewScriptCookies(s.Response.Headers, s.SentRequest.Host())
}

func (s *SendExecution) AddTests(tests []TestResult) {
	for _, test := range tests {
		test.Index = len(s.Tests)
		s.Tests = append(s.Tests, test)
	}
}

func (s *SendExecution) TestSummary() TestSummary {
	return NewTestSummary(s.Tests)
}

func (s *SendExecution) ScriptErrors() []ScriptError {
	errors := make([]ScriptError, 0)
	for _, run := range append(append([]ScriptRun{}, s.PreRuns...), s.PostRuns...) {
		if run.Error != nil {
			errors = append(errors, *run.Error)
		}
	}
	return errors
}

func (s *SendExecution) AddNested(nested NestedRequest) {
	s.Nested = append(s.Nested, nested)
}

func (s SendExecution) LogValue() any {
	value := map[string]any{"id": s.ID, "outcome": s.Outcome, "pre_runs": len(s.PreRuns), "post_runs": len(s.PostRuns)}
	if s.Response != nil {
		value["status_code"] = s.Response.StatusCode
	}
	return value
}
