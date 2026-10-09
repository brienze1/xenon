package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
)

type Body struct {
	Raw        BodyRaw       `json:"raw"`
	GraphQl    BodyGraphQl   `json:"graphql"`
	Binary     BodyBinary    `json:"binary"`
	Mode       string        `json:"mode" validate:"omitempty,oneof=none formdata urlencoded raw binary graphql"`
	FormData   []FormDataRow `json:"formdata" validate:"dive"`
	UrlEncoded []KeyValue    `json:"urlencoded"`
}

func NewBody(body entity.Body) Body {
	result := Body{
		Mode:       body.Mode.String(),
		Raw:        NewBodyRaw(body.Raw),
		GraphQl:    NewBodyGraphQl(body.GraphQl),
		Binary:     NewBodyBinary(body.Binary),
		FormData:   make([]FormDataRow, 0, len(body.FormData)),
		UrlEncoded: NewKeyValues(body.UrlEncoded),
	}
	for _, row := range body.FormData {
		result.FormData = append(result.FormData, NewFormDataRow(row))
	}
	return result
}

func (b *Body) ToEntity() entity.Body {
	body := entity.DefaultBody()
	if b.Mode != "" {
		body.Mode = bodymode.GetBodyMode(b.Mode)
	}
	body.Raw = b.Raw.ToEntity()
	body.GraphQl = b.GraphQl.ToEntity()
	body.Binary = b.Binary.ToEntity()
	for index := range b.FormData {
		body.FormData = append(body.FormData, b.FormData[index].ToEntity())
	}
	body.UrlEncoded = keyValuesToEntity(b.UrlEncoded)
	return body
}
