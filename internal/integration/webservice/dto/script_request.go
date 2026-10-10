package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

const (
	scriptModeRaw        = "raw"
	scriptModeUrlEncoded = "urlencoded"
	scriptModeFormData   = "formdata"
	scriptModeGraphQl    = "graphql"
	scriptAuthToken      = "token"
	scriptAuthUsername   = "username"
	scriptAuthPassword   = "password"
)

type ScriptKeyValue struct {
	Value    any    `json:"value"`
	Key      string `json:"key"`
	Disabled bool   `json:"disabled,omitempty"`
}

type ScriptGraphQl struct {
	Variables     any    `json:"variables,omitempty"`
	Query         string `json:"query"`
	OperationName string `json:"operationName,omitempty"`
}

type ScriptBody struct {
	GraphQl    *ScriptGraphQl   `json:"graphql,omitempty"`
	Mode       string           `json:"mode,omitempty"`
	Raw        string           `json:"raw,omitempty"`
	UrlEncoded []ScriptKeyValue `json:"urlencoded,omitempty"`
	FormData   []ScriptKeyValue `json:"formdata,omitempty"`
}

type ScriptAuth struct {
	Type   string           `json:"type"`
	Bearer []ScriptKeyValue `json:"bearer,omitempty"`
	Basic  []ScriptKeyValue `json:"basic,omitempty"`
}

// ScriptRequest is the request as pm.request sees it (Postman's request JSON shape).
type ScriptRequest struct {
	Auth   *ScriptAuth      `json:"auth,omitempty"`
	Body   *ScriptBody      `json:"body,omitempty"`
	Method string           `json:"method"`
	URL    string           `json:"url"`
	Header []ScriptKeyValue `json:"header"`
}

func NewScriptRequest(request entity.Request) ScriptRequest {
	script := ScriptRequest{Method: request.MethodOrDefault(), URL: request.URL, Header: newScriptKeyValues(request.Headers), Auth: newScriptAuth(request.Auth)}
	if body := newScriptBody(request.Body); body.Mode != "" {
		script.Body = &body
	}
	return script
}

func newScriptAuth(auth entity.Auth) *ScriptAuth {
	switch auth.Type {
	case authtype.Bearer:
		return &ScriptAuth{Type: auth.Type.String(), Bearer: []ScriptKeyValue{{Key: scriptAuthToken, Value: auth.Bearer.Token}}}
	case authtype.Basic:
		return &ScriptAuth{Type: auth.Type.String(), Basic: []ScriptKeyValue{{Key: scriptAuthUsername, Value: auth.Basic.Username}, {Key: scriptAuthPassword, Value: auth.Basic.Password}}}
	case authtype.None:
		return &ScriptAuth{Type: auth.Type.String()}
	default:
		return nil
	}
}

func newScriptBody(body entity.Body) ScriptBody {
	switch body.Mode {
	case bodymode.Raw:
		return ScriptBody{Mode: scriptModeRaw, Raw: body.Raw.Content}
	case bodymode.UrlEncoded:
		return ScriptBody{Mode: scriptModeUrlEncoded, UrlEncoded: newScriptKeyValues(body.UrlEncoded)}
	case bodymode.FormData:
		rows := make([]ScriptKeyValue, 0, len(body.FormData))
		for _, row := range body.FormData {
			rows = append(rows, ScriptKeyValue{Key: row.Key, Value: row.Value, Disabled: !row.Enabled})
		}
		return ScriptBody{Mode: scriptModeFormData, FormData: rows}
	case bodymode.GraphQl:
		return ScriptBody{Mode: scriptModeGraphQl, GraphQl: &ScriptGraphQl{Query: body.GraphQl.Query, Variables: optionalText(body.GraphQl.Variables)}}
	default:
		return ScriptBody{}
	}
}

func newScriptKeyValues(rows []entity.KeyValue) []ScriptKeyValue {
	values := make([]ScriptKeyValue, 0, len(rows))
	for _, row := range rows {
		values = append(values, ScriptKeyValue{Key: row.Key, Value: row.Value, Disabled: !row.Enabled})
	}
	return values
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
	if !body.Equals(newScriptBody(written.Body)) {
		request.Body = body.toEntity(written.Body)
	}
	if s.Auth != nil && !s.Auth.Equals(newScriptAuth(written.Auth)) {
		request.Auth = s.Auth.toEntity(written.Auth)
	}
	return request
}

func (s ScriptBody) Equals(other ScriptBody) bool {
	return s.Mode == other.Mode && s.Raw == other.Raw && sameKeyValues(s.UrlEncoded, other.UrlEncoded) && sameKeyValues(s.FormData, other.FormData) && s.GraphQl.Equals(other.GraphQl)
}

func (s *ScriptGraphQl) Equals(other *ScriptGraphQl) bool {
	if s == nil || other == nil {
		return s == other
	}
	return s.Query == other.Query && s.OperationName == other.OperationName && scriptText(s.Variables) == scriptText(other.Variables)
}

func (s *ScriptAuth) Equals(other *ScriptAuth) bool {
	if other == nil {
		return false
	}
	return s.Type == other.Type && sameKeyValues(s.Bearer, other.Bearer) && sameKeyValues(s.Basic, other.Basic)
}

func (s *ScriptAuth) toEntity(written entity.Auth) entity.Auth {
	auth := written
	auth.Type = authtype.GetAuthType(s.Type)
	switch auth.Type {
	case authtype.Bearer:
		auth.Bearer = entity.AuthBearer{Token: scriptParameter(s.Bearer, scriptAuthToken)}
	case authtype.Basic:
		auth.Basic = entity.AuthBasic{Username: scriptParameter(s.Basic, scriptAuthUsername), Password: scriptParameter(s.Basic, scriptAuthPassword)}
	case authtype.None:
	default:
		return written
	}
	return auth
}

func scriptParameter(values []ScriptKeyValue, key string) string {
	for _, value := range values {
		if value.Key == key {
			return scriptText(value.Value)
		}
	}
	return ""
}

func optionalText(text string) any {
	if text == "" {
		return nil
	}
	return text
}

func (s ScriptBody) toEntity(written entity.Body) entity.Body {
	body := entity.DefaultBody()
	switch s.Mode {
	case scriptModeRaw:
		body.Mode = bodymode.Raw
		body.Raw = entity.BodyRaw{Language: rawLanguageOf(written), Content: s.Raw}
	case scriptModeUrlEncoded:
		body.Mode = bodymode.UrlEncoded
		body.UrlEncoded = scriptKeyValuesToEntity(s.UrlEncoded)
	case scriptModeFormData:
		body.Mode = bodymode.FormData
		for _, row := range scriptKeyValuesToEntity(s.FormData) {
			formRow := entity.NewFormDataRow(row.Key, row.Value)
			formRow.Enabled = row.Enabled
			body.FormData = append(body.FormData, formRow)
		}
	case scriptModeGraphQl:
		body.Mode = bodymode.GraphQl
		if s.GraphQl != nil {
			body.GraphQl = entity.BodyGraphQl{Query: s.GraphQl.Query, Variables: scriptText(s.GraphQl.Variables)}
		}
	}
	return body
}

func rawLanguageOf(written entity.Body) enums.RawLanguage {
	if written.Mode == bodymode.Raw {
		return written.Raw.Language
	}
	return rawlanguage.Text
}

func scriptKeyValuesToEntity(values []ScriptKeyValue) []entity.KeyValue {
	rows := make([]entity.KeyValue, 0, len(values))
	for _, value := range values {
		rows = append(rows, entity.KeyValue{Key: value.Key, Value: scriptText(value.Value), Enabled: !value.Disabled})
	}
	return rows
}

func sameKeyValues(a, b []ScriptKeyValue) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index].Key != b[index].Key || scriptText(a[index].Value) != scriptText(b[index].Value) || a[index].Disabled != b[index].Disabled {
			return false
		}
	}
	return true
}

func scriptText(value any) string {
	if value == nil {
		return ""
	}
	return entity.LocalVariableText(value)
}
