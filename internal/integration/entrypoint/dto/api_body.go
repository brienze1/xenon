package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiBody struct {
	MediaType string     `json:"media_type"`
	Example   string     `json:"example"`
	Fields    []ApiField `json:"fields"`
	Binary    bool       `json:"binary"`
}

func NewApiBody(body entity.ApiBody) ApiBody {
	documented := ApiBody{MediaType: body.MediaType, Example: body.Example, Fields: make([]ApiField, 0, len(body.Fields)), Binary: body.Binary}
	for _, field := range body.Fields {
		documented.Fields = append(documented.Fields, NewApiField(field))
	}
	return documented
}
