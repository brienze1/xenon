package dto

import "github.com/brienze1/xenon/internal/domain/entity"

// ScriptRequest is the request as pm.request sees it (Postman's request JSON shape).
type ScriptRequest struct {
	Auth   *ScriptAuth      `json:"auth,omitempty"`
	Body   *ScriptBody      `json:"body,omitempty"`
	Method string           `json:"method"`
	URL    string           `json:"url"`
	Header []ScriptKeyValue `json:"header"`
}

func NewScriptRequest(request entity.Request) ScriptRequest {
	script := ScriptRequest{Method: request.MethodOrDefault(), URL: request.URL, Header: NewScriptKeyValues(request.Headers), Auth: NewScriptAuth(request.Auth)}
	if body := NewScriptBody(request.Body); body.Mode != "" {
		script.Body = &body
	}
	return script
}

// ToEntity applies the script's view to the request as written; a body the script did not change keeps its original form.
func (s ScriptRequest) ToEntity(written entity.Request) entity.Request {
	request := written
	request.Method = s.Method
	request.URL = s.URL
	request.Headers = scriptKeyValuesToEntity(s.Header)
	body := ScriptBody{}
	if s.Body != nil {
		body = *s.Body
	}
	if !body.Equals(NewScriptBody(written.Body)) {
		request.Body = body.ToEntity(written.Body)
	}
	if s.Auth != nil && !s.Auth.Equals(NewScriptAuth(written.Auth)) {
		request.Auth = s.Auth.ToEntity(written.Auth)
	}
	return request
}
