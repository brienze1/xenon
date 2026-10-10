package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Body struct {
	Raw        *BodyRaw      `yaml:"raw" json:"raw,omitempty"`
	Binary     *BodyBinary   `yaml:"binary" json:"binary,omitempty"`
	GraphQl    *BodyGraphQl  `yaml:"graphql" json:"graphql,omitempty"`
	Mode       string        `yaml:"mode" json:"mode"`
	FormData   []FormDataRow `yaml:"formdata" json:"formdata"`
	UrlEncoded []KeyValue    `yaml:"urlencoded" json:"urlencoded"`
}

func NewBody(body entity.Body) Body {
	model := Body{Mode: body.Mode.String(), UrlEncoded: newKeyValues(body.UrlEncoded), FormData: make([]FormDataRow, 0, len(body.FormData))}
	for _, row := range body.FormData {
		model.FormData = append(model.FormData, NewFormDataRow(row))
	}
	if body.HasRaw() {
		raw := NewBodyRaw(body.Raw)
		model.Raw = &raw
	}
	if body.HasBinary() {
		binary := NewBodyBinary(body.Binary)
		model.Binary = &binary
	}
	if body.HasGraphQl() {
		graphQl := NewBodyGraphQl(body.GraphQl)
		model.GraphQl = &graphQl
	}
	return model
}

func (b *Body) ToEntity() entity.Body {
	body := entity.DefaultBody()
	if b.Mode != "" {
		body.Mode = bodymode.GetBodyMode(b.Mode)
	}
	for index := range b.FormData {
		body.FormData = append(body.FormData, b.FormData[index].ToEntity())
	}
	body.UrlEncoded = keyValuesToEntity(b.UrlEncoded)
	if b.Raw != nil {
		body.Raw = b.Raw.ToEntity()
	}
	if b.Binary != nil {
		body.Binary = b.Binary.ToEntity()
	}
	if b.GraphQl != nil {
		body.GraphQl = b.GraphQl.ToEntity()
	}
	return body
}

func (b *Body) Field() yamlnode.Field {
	fields := []yamlnode.Field{yamlnode.String("mode", b.ToEntity().Mode.String())}
	if len(b.FormData) > 0 {
		fields = append(fields, yamlnode.List("formdata", b.formDataNodes()))
	}
	if len(b.UrlEncoded) > 0 {
		fields = append(fields, yamlnode.List("urlencoded", keyValueNodes(b.UrlEncoded)))
	}
	if b.Raw != nil {
		fields = append(fields, yamlnode.Object("raw", yamlnode.String("language", b.Raw.Language), yamlnode.String("content", b.Raw.Content)))
	}
	if b.Binary != nil {
		fields = append(fields, yamlnode.Object("binary", yamlnode.String("src", b.Binary.Src), fileRefStateField(b.Binary.ToEntity())))
	}
	if b.GraphQl != nil {
		fields = append(fields, yamlnode.Object("graphql", yamlnode.String("query", b.GraphQl.Query), yamlnode.String("variables", b.GraphQl.Variables)))
	}
	return yamlnode.Object("body", fields...)
}

func (b *Body) formDataNodes() []*yaml.Node {
	nodes := make([]*yaml.Node, 0, len(b.FormData))
	for index := range b.FormData {
		nodes = append(nodes, b.FormData[index].Node())
	}
	return nodes
}
