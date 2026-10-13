package dto

import (
	"encoding/json"
	"slices"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

const (
	postmanFileMode    = "file"
	postmanRawLanguage = "language"
)

var (
	postmanFormData     = bodymode.FormData.String()
	postmanUrlEncoded   = bodymode.UrlEncoded.String()
	postmanRaw          = bodymode.Raw.String()
	postmanGraphQl      = bodymode.GraphQl.String()
	postmanRawLanguages = []string{rawlanguage.Json.String(), rawlanguage.JavaScript.String(), rawlanguage.Xml.String(), rawlanguage.Html.String(), rawlanguage.Text.String()}
)

type Body struct {
	Raw        map[string]any `json:"-"`
	File       *BodyFile      `json:"file"`
	GraphQl    *BodyGraphQl   `json:"graphql"`
	Mode       string         `json:"mode"`
	Content    string         `json:"raw"`
	UrlEncoded []KeyValue     `json:"urlencoded"`
	FormData   []FormDataRow  `json:"formdata"`
	Disabled   bool           `json:"disabled"`
}

func (b *Body) UnmarshalJSON(data []byte) error {
	type plain Body
	if err := json.Unmarshal(data, (*plain)(b)); err != nil {
		return err
	}
	b.Raw = decodeRaw(data)
	return nil
}

func (b *Body) ToEntity() entity.Body {
	body := entity.DefaultBody()
	if b == nil || b.Disabled {
		return body
	}
	switch b.Mode {
	case postmanRaw:
		body.Mode = bodymode.Raw
		body.Raw = entity.NewBodyRaw(b.language(), b.Content)
	case postmanUrlEncoded:
		body.Mode = bodymode.UrlEncoded
		body.UrlEncoded = keyValuesToEntity(b.UrlEncoded)
	case postmanFormData:
		body.Mode = bodymode.FormData
		for index := range b.FormData {
			body.FormData = append(body.FormData, b.FormData[index].ToEntity())
		}
	case postmanFileMode:
		body.Mode = bodymode.Binary
		body.Binary = entity.NewRepickFileRef(b.fileSrc())
	case postmanGraphQl:
		body.Mode = bodymode.GraphQl
		body.GraphQl = entity.BodyGraphQl{Query: b.GraphQl.Query, Variables: text(b.GraphQl.Variables)}
	}
	return body
}

func (b *Body) FileRefs() []string {
	refs := make([]string, 0)
	if b == nil || b.Disabled {
		return refs
	}
	if b.Mode == postmanFileMode {
		refs = append(refs, b.fileSrc())
	}
	if b.Mode == postmanFormData {
		for index := range b.FormData {
			if b.FormData[index].IsFile() {
				refs = append(refs, b.FormData[index].firstSrc())
			}
		}
	}
	return refs
}

func (b *Body) Passthrough() map[string]any {
	if b.Disabled {
		return b.Raw
	}
	fields := without(b.Raw, "mode", "raw", "urlencoded", "formdata", "file", "graphql", "options")
	fields = withEntry(fields, "options", b.options())
	fields = withList(fields, "urlencoded", keyValuePassthrough(b.UrlEncoded))
	rows := make([]any, 0, len(b.FormData))
	for index := range b.FormData {
		rows = append(rows, b.FormData[index].Passthrough())
	}
	fields = withList(fields, "formdata", rows)
	if file, isMap := b.Raw["file"].(map[string]any); isMap {
		fields = withEntry(fields, "file", without(file, "src"))
	}
	return fields
}

func (b *Body) language() string {
	if options, isMap := b.Raw["options"].(map[string]any); isMap {
		if raw, isMap := options[postmanRaw].(map[string]any); isMap && slices.Contains(postmanRawLanguages, text(raw[postmanRawLanguage])) {
			return text(raw[postmanRawLanguage])
		}
	}
	return rawlanguage.Text.String()
}

func (b *Body) options() map[string]any {
	options, _ := b.Raw["options"].(map[string]any)
	rest := without(options, postmanRaw)
	if raw, isMap := options[postmanRaw].(map[string]any); isMap {
		rest = withEntry(rest, postmanRaw, without(raw, postmanRawLanguage))
	}
	return rest
}

func (b *Body) fileSrc() string {
	if b.File == nil {
		return ""
	}
	return text(b.File.Src)
}

func NewBody(body entity.Body, passthrough any) *Body {
	fields := asMap(passthrough)
	if fields["disabled"] == true {
		if body.HasContent() {
			fields = map[string]any{}
		} else {
			return &Body{Raw: fields, Disabled: true}
		}
	}
	exported := &Body{Raw: without(fields, "urlencoded", "formdata", "file", "disabled")}
	switch body.Mode {
	case bodymode.Raw:
		exported.Mode, exported.Content = postmanRaw, body.Raw.Content
		exported.Raw["options"] = rawOptions(body.Raw.Language.String(), asMap(fields["options"]))
	case bodymode.UrlEncoded:
		exported.Mode = postmanUrlEncoded
		for index, row := range body.UrlEncoded {
			exported.UrlEncoded = append(exported.UrlEncoded, NewKeyValueRow(row, mapAt(asList(fields["urlencoded"]), index)))
		}
	case bodymode.FormData:
		exported.Mode = postmanFormData
		for index, row := range body.FormData {
			exported.FormData = append(exported.FormData, NewFormDataRowOf(row, mapAt(asList(fields["formdata"]), index)))
		}
	case bodymode.Binary:
		exported.Mode, exported.File = postmanFileMode, &BodyFile{Src: body.Binary.Src}
		exported.Raw["file"] = without(asMap(fields["file"]), "src")
	case bodymode.GraphQl:
		exported.Mode, exported.GraphQl = postmanGraphQl, &BodyGraphQl{Query: body.GraphQl.Query, Variables: body.GraphQl.Variables}
	default:
		return nil
	}
	return exported
}

func (b Body) MarshalJSON() ([]byte, error) {
	if b.Mode == "" {
		return encodeJson(b.Raw)
	}
	body := newObject().set("mode", b.Mode)
	switch b.Mode {
	case postmanRaw:
		body.set("raw", b.Content)
	case postmanUrlEncoded:
		body.set("urlencoded", nonNil(b.UrlEncoded))
	case postmanFormData:
		body.set("formdata", nonNil(b.FormData))
	case postmanFileMode:
		body.set("file", newObject().set("src", b.File.Src).merged(asMap(b.Raw["file"])))
	case postmanGraphQl:
		body.set("graphql", newObject().set("query", b.GraphQl.Query).set("variables", b.GraphQl.Variables))
	}
	options := asMap(b.Raw["options"])
	body.setIf(len(options) > 0, "options", options)
	return body.merged(without(b.Raw, "file", "options")).MarshalJSON()
}

func rawOptions(language string, passthrough map[string]any) map[string]any {
	options := without(passthrough, postmanRaw)
	raw := without(asMap(passthrough[postmanRaw]), postmanRawLanguage)
	if language != rawlanguage.Text.String() {
		raw[postmanRawLanguage] = language
	}
	if len(raw) > 0 {
		options[postmanRaw] = raw
	}
	return options
}

func nonNil[T any](rows []T) []T {
	if rows == nil {
		return []T{}
	}
	return rows
}
