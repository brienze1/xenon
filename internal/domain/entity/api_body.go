package entity

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

const (
	MediaTypeJson        = "application/json"
	MediaTypeUrlEncoded  = "application/x-www-form-urlencoded"
	MediaTypeMultipart   = "multipart/form-data"
	MediaTypeOctetStream = "application/octet-stream"
	MediaTypeText        = "text/plain"
)

type ApiBody struct {
	MediaType string
	Example   string
	Fields    []ApiField
	Binary    bool
}

func NewApiBody(body Body) ApiBody {
	switch body.Mode {
	case bodymode.Raw:
		return ApiBody{MediaType: body.Raw.ContentType(), Example: body.Raw.Content}
	case bodymode.UrlEncoded:
		return ApiBody{MediaType: MediaTypeUrlEncoded, Fields: urlEncodedFields(body.UrlEncoded)}
	case bodymode.FormData:
		return ApiBody{MediaType: MediaTypeMultipart, Fields: formDataFields(body.FormData)}
	case bodymode.Binary:
		return ApiBody{MediaType: MediaTypeOctetStream, Binary: true}
	case bodymode.GraphQl:
		return ApiBody{MediaType: MediaTypeJson, Example: graphQlExample(body.GraphQl)}
	default:
		return ApiBody{}
	}
}

func (a ApiBody) HasContent() bool {
	return a.MediaType != ""
}

func (a ApiBody) IsForm() bool {
	return a.MediaType == MediaTypeUrlEncoded || a.MediaType == MediaTypeMultipart
}

func (a ApiBody) Equal(other ApiBody) bool {
	return a.MediaType == other.MediaType && a.Example == other.Example && a.Binary == other.Binary && slices.Equal(a.Fields, other.Fields)
}

// AppliedTo turns the documented body into the request body, keeping what OpenAPI cannot express.
func (a ApiBody) AppliedTo(current Body) Body {
	switch {
	case !a.HasContent():
		current.Mode = bodymode.None
	case a.MediaType == MediaTypeUrlEncoded:
		current.Mode = bodymode.UrlEncoded
		current.UrlEncoded = append(a.keyValues(), disabledKeyValues(current.UrlEncoded)...)
	case a.MediaType == MediaTypeMultipart:
		current.Mode = bodymode.FormData
		current.FormData = append(a.formDataRows(current.FormData), disabledFormDataRows(current.FormData)...)
	case a.Binary:
		current.Mode = bodymode.Binary
	case current.Mode == bodymode.GraphQl && a.isGraphQl():
		current.GraphQl = a.graphQl()
	default:
		current.Mode = bodymode.Raw
		current.Raw = BodyRaw{Language: RawLanguageOf(a.MediaType), Content: a.Example}
	}
	return current
}

func RawLanguageOf(mediaType string) enums.RawLanguage {
	lower := strings.ToLower(mediaType)
	switch {
	case strings.Contains(lower, "json"):
		return rawlanguage.Json
	case strings.Contains(lower, "xml"):
		return rawlanguage.Xml
	case strings.Contains(lower, "html"):
		return rawlanguage.Html
	case strings.Contains(lower, "javascript"):
		return rawlanguage.JavaScript
	default:
		return rawlanguage.Text
	}
}

func (a ApiBody) keyValues() []KeyValue {
	rows := make([]KeyValue, 0, len(a.Fields))
	for _, field := range a.Fields {
		rows = append(rows, field.KeyValue())
	}
	return rows
}

func (a ApiBody) formDataRows(current []FormDataRow) []FormDataRow {
	rows := make([]FormDataRow, 0, len(a.Fields))
	for _, field := range a.Fields {
		rows = append(rows, field.FormDataRow(current))
	}
	return rows
}

func (a ApiBody) isGraphQl() bool {
	var document map[string]any
	if json.Unmarshal([]byte(a.Example), &document) != nil {
		return false
	}
	_, hasQuery := document["query"].(string)
	return hasQuery
}

func (a ApiBody) graphQl() BodyGraphQl {
	var document struct {
		Query     string          `json:"query"`
		Variables json.RawMessage `json:"variables"`
	}
	_ = json.Unmarshal([]byte(a.Example), &document)
	graphQl := BodyGraphQl{Query: document.Query}
	if len(document.Variables) > 0 && string(document.Variables) != "null" {
		graphQl.Variables = string(document.Variables)
	}
	return graphQl
}

func graphQlExample(graphQl BodyGraphQl) string {
	document := map[string]any{"query": graphQl.Query}
	var variables any
	if strings.TrimSpace(graphQl.Variables) != "" && json.Unmarshal([]byte(graphQl.Variables), &variables) == nil {
		document["variables"] = variables
	}
	encoded, _ := json.Marshal(document)
	return string(encoded)
}

func urlEncodedFields(rows []KeyValue) []ApiField {
	fields := make([]ApiField, 0, len(rows))
	for _, row := range rows {
		if row.IsActive() {
			fields = append(fields, NewApiField(row))
		}
	}
	return fields
}

func formDataFields(rows []FormDataRow) []ApiField {
	fields := make([]ApiField, 0, len(rows))
	for _, row := range rows {
		if row.IsActive() {
			fields = append(fields, NewApiFormField(row))
		}
	}
	return fields
}

func disabledKeyValues(rows []KeyValue) []KeyValue {
	return slices.DeleteFunc(slices.Clone(rows), KeyValue.IsActive)
}

func disabledFormDataRows(rows []FormDataRow) []FormDataRow {
	return slices.DeleteFunc(slices.Clone(rows), FormDataRow.IsActive)
}
