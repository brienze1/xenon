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
