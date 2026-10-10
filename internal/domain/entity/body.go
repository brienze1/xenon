package entity

import (
	"strconv"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

type Body struct {
	Raw        BodyRaw
	GraphQl    BodyGraphQl
	Binary     FileRef
	Mode       enums.BodyMode
	FormData   []FormDataRow
	UrlEncoded []KeyValue
}

func DefaultBody() Body {
	return Body{
		Mode:       bodymode.None,
		Raw:        BodyRaw{Language: rawlanguage.Json},
		FormData:   []FormDataRow{},
		UrlEncoded: []KeyValue{},
	}
}

func (b Body) HasContent() bool {
	return b.Mode != bodymode.None
}

func (b Body) ContentType() string {
	switch b.Mode {
	case bodymode.Raw:
		return b.Raw.ContentType()
	case bodymode.UrlEncoded:
		return "application/x-www-form-urlencoded"
	case bodymode.FormData:
		return "multipart/form-data; boundary=" + ComputedAtSend
	case bodymode.GraphQl:
		return "application/json"
	case bodymode.Binary:
		return "application/octet-stream"
	default:
		return ""
	}
}

func (b Body) MapText(transform func(string) string) Body {
	b.Raw = b.Raw.MapText(transform)
	b.GraphQl = b.GraphQl.MapText(transform)
	b.Binary = b.Binary.WithSrc(transform(b.Binary.Src))
	rows := make([]FormDataRow, 0, len(b.FormData))
	for _, row := range b.FormData {
		rows = append(rows, row.MapText(transform))
	}
	b.FormData = rows
	b.UrlEncoded = mapKeyValues(b.UrlEncoded, transform)
	return b
}

func (b Body) TextFields(prefix string) []TextField {
	switch b.Mode {
	case bodymode.Raw:
		return []TextField{NewTextField(fieldName(prefix, "raw"), b.Raw.Content)}
	case bodymode.GraphQl:
		return []TextField{NewTextField(fieldName(prefix, "graphql", "query"), b.GraphQl.Query), NewTextField(fieldName(prefix, "graphql", "variables"), b.GraphQl.Variables)}
	case bodymode.Binary:
		return []TextField{NewTextField(fieldName(prefix, "binary"), b.Binary.Src)}
	case bodymode.FormData:
		return b.formDataFields(fieldName(prefix, "formdata"))
	case bodymode.UrlEncoded:
		return keyValueFields(fieldName(prefix, "urlencoded"), b.UrlEncoded)
	default:
		return []TextField{}
	}
}

func (b Body) formDataFields(prefix string) []TextField {
	fields := make([]TextField, 0, len(b.FormData))
	for index, row := range b.FormData {
		position := strconv.Itoa(index)
		fields = append(fields, NewTextField(fieldName(prefix, position, "key"), row.Key))
		if row.IsFile() {
			fields = append(fields, NewTextField(fieldName(prefix, position, "src"), row.Src.Src))
			continue
		}
		fields = append(fields, NewTextField(fieldName(prefix, position, "value"), row.Value))
	}
	return fields
}

func (b Body) IsMultipart() bool {
	return b.Mode == bodymode.FormData
}

func (b Body) ResolveFiles(workspaceRoot string) Body {
	b.Binary = b.Binary.WithSrc(b.Binary.Resolve(workspaceRoot))
	rows := make([]FormDataRow, 0, len(b.FormData))
	for _, row := range b.FormData {
		row.Src = row.Src.WithSrc(row.Src.Resolve(workspaceRoot))
		rows = append(rows, row)
	}
	b.FormData = rows
	return b
}

func (b Body) RelativeTo(workspaceRoot string) Body {
	b.Binary = b.Binary.RelativeTo(workspaceRoot)
	rows := make([]FormDataRow, 0, len(b.FormData))
	for _, row := range b.FormData {
		row.Src = row.Src.RelativeTo(workspaceRoot)
		rows = append(rows, row)
	}
	b.FormData = rows
	return b
}

func (b Body) FileRefs() []FileRef {
	refs := make([]FileRef, 0)
	switch b.Mode {
	case bodymode.Binary:
		refs = append(refs, b.Binary)
	case bodymode.FormData:
		for _, row := range b.FormData {
			if row.IsActive() && row.IsFile() {
				refs = append(refs, row.Src)
			}
		}
	}
	return refs
}

func (b Body) HasRaw() bool {
	return b.Mode == bodymode.Raw || b.Raw.HasContent()
}

func (b Body) HasBinary() bool {
	return b.Mode == bodymode.Binary || !b.Binary.IsEmpty()
}

func (b Body) HasGraphQl() bool {
	return b.Mode == bodymode.GraphQl || !b.GraphQl.IsEmpty()
}

func (b Body) LogValue() any {
	return map[string]any{"mode": b.Mode, "formData": len(b.FormData), "urlEncoded": len(b.UrlEncoded)}
}
