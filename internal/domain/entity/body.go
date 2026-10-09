package entity

import (
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

func (b Body) IsMultipart() bool {
	return b.Mode == bodymode.FormData
}

func (b Body) ResolveFiles(workspaceRoot string) Body {
	b.Binary = FileRef{Src: b.Binary.Resolve(workspaceRoot)}
	rows := make([]FormDataRow, 0, len(b.FormData))
	for _, row := range b.FormData {
		row.Src = FileRef{Src: row.Src.Resolve(workspaceRoot)}
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
			if row.Enabled && row.IsFile() {
				refs = append(refs, row.Src)
			}
		}
	}
	return refs
}
