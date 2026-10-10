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
