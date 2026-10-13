package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

const (
	JsonBody           = "json"
	TextBody           = "text"
	XmlBody            = "xml"
	SparqlBody         = "sparql"
	FormUrlEncodedBody = "formUrlEncoded"
	MultipartFormBody  = "multipartForm"
	GraphQlBody        = "graphql"
	FileBody           = "file"
)

var rawBodyLanguages = map[string]string{
	JsonBody:   rawlanguage.Json.String(),
	TextBody:   rawlanguage.Text.String(),
	XmlBody:    rawlanguage.Xml.String(),
	SparqlBody: rawlanguage.Text.String(),
}

type Body struct {
	Mode      string
	Text      string
	Query     string
	Variables string
	Pairs     []Pair
	Multipart []MultipartRow
	Files     []FilePart
}

func NewBody(mode string) Body {
	return Body{Mode: mode, Pairs: []Pair{}, Multipart: []MultipartRow{}, Files: []FilePart{}}
}

func (b Body) ToEntity() entity.Body {
	body := entity.DefaultBody()
	if language, isRaw := rawBodyLanguages[b.Mode]; isRaw {
		body.Mode = bodymode.Raw
		body.Raw = entity.NewBodyRaw(language, b.Text)
		return body
	}
	switch b.Mode {
	case FormUrlEncodedBody:
		body.Mode = bodymode.UrlEncoded
		body.UrlEncoded = keyValues(b.Pairs)
	case MultipartFormBody:
		body.Mode = bodymode.FormData
		for _, row := range b.Multipart {
			body.FormData = append(body.FormData, row.toEntity())
		}
	case GraphQlBody:
		body.Mode = bodymode.GraphQl
		body.GraphQl = entity.BodyGraphQl{Query: b.Query, Variables: b.Variables}
	case FileBody:
		body.Mode = bodymode.Binary
		body.Binary = entity.NewRepickFileRef(b.selectedFile().Path)
	}
	return body
}

func (b Body) Passthrough() map[string]any {
	fields := map[string]any{}
	if b.Mode == SparqlBody {
		fields["bodyMode"] = SparqlBody
	}
	contentTypes := make([]any, 0)
	for _, row := range b.Multipart {
		if row.ContentType != "" {
			contentTypes = append(contentTypes, map[string]any{"key": row.Key, "contentType": row.ContentType})
		}
	}
	if len(contentTypes) > 0 {
		fields["formdata"] = contentTypes
	}
	if file := b.selectedFile(); b.Mode == FileBody && file.ContentType != "" {
		fields["file"] = map[string]any{"contentType": file.ContentType}
	}
	return fields
}

func (b Body) selectedFile() FilePart {
	for _, file := range b.Files {
		if file.Selected {
			return file
		}
	}
	if len(b.Files) > 0 {
		return b.Files[0]
	}
	return NewFilePart("", "", false)
}
