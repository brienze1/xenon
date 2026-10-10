package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

const (
	scriptModeRaw        = "raw"
	scriptModeUrlEncoded = "urlencoded"
	scriptModeFormData   = "formdata"
)

type ScriptKeyValue struct {
	Value    any    `json:"value"`
	Key      string `json:"key"`
	Disabled bool   `json:"disabled,omitempty"`
}

type ScriptBody struct {
	Mode       string           `json:"mode,omitempty"`
	Raw        string           `json:"raw,omitempty"`
	UrlEncoded []ScriptKeyValue `json:"urlencoded,omitempty"`
	FormData   []ScriptKeyValue `json:"formdata,omitempty"`
}

// ScriptRequest is the request as pm.request sees it (Postman's request JSON shape).
type ScriptRequest struct {
	Body   ScriptBody       `json:"body"`
	Method string           `json:"method"`
	URL    string           `json:"url"`
	Header []ScriptKeyValue `json:"header"`
}

func NewScriptRequest(request entity.Request) ScriptRequest {
	return ScriptRequest{Method: request.MethodOrDefault(), URL: request.URL, Header: newScriptKeyValues(request.Headers), Body: newScriptBody(request.Body)}
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
	if !s.Body.Equals(newScriptBody(written.Body)) {
		request.Body = s.Body.toEntity(written.Body)
	}
	return request
}

func (s ScriptBody) Equals(other ScriptBody) bool {
	return s.Mode == other.Mode && s.Raw == other.Raw && sameKeyValues(s.UrlEncoded, other.UrlEncoded) && sameKeyValues(s.FormData, other.FormData)
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
