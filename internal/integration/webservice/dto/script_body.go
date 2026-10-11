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
	scriptModeGraphQl    = "graphql"
)

type ScriptBody struct {
	GraphQl    *ScriptGraphQl   `json:"graphql,omitempty"`
	Mode       string           `json:"mode,omitempty"`
	Raw        string           `json:"raw,omitempty"`
	UrlEncoded []ScriptKeyValue `json:"urlencoded,omitempty"`
	FormData   []ScriptKeyValue `json:"formdata,omitempty"`
}

func NewScriptBody(body entity.Body) ScriptBody {
	switch body.Mode {
	case bodymode.Raw:
		return ScriptBody{Mode: scriptModeRaw, Raw: body.Raw.Content}
	case bodymode.UrlEncoded:
		return ScriptBody{Mode: scriptModeUrlEncoded, UrlEncoded: NewScriptKeyValues(body.UrlEncoded)}
	case bodymode.FormData:
		rows := make([]ScriptKeyValue, 0, len(body.FormData))
		for _, row := range body.FormData {
			rows = append(rows, NewScriptKeyValue(entity.KeyValue{Key: row.Key, Value: row.Value, Enabled: row.Enabled}))
		}
		return ScriptBody{Mode: scriptModeFormData, FormData: rows}
	case bodymode.GraphQl:
		graphQl := NewScriptGraphQl(body.GraphQl)
		return ScriptBody{Mode: scriptModeGraphQl, GraphQl: &graphQl}
	default:
		return ScriptBody{}
	}
}

func (s ScriptBody) Equals(other ScriptBody) bool {
	return s.Mode == other.Mode && s.Raw == other.Raw && sameKeyValues(s.UrlEncoded, other.UrlEncoded) && sameKeyValues(s.FormData, other.FormData) && s.GraphQl.Equals(other.GraphQl)
}

func (s ScriptBody) ToEntity(written entity.Body) entity.Body {
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
			body.GraphQl = s.GraphQl.ToEntity()
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
