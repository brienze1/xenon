package dto

import (
	"bytes"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	openApiVersion = "3.1.0"
	yamlIndent     = 2
)

type Document struct {
	Components *Components `yaml:"components,omitempty"`
	Info       Info        `yaml:"info"`
	OpenApi    string      `yaml:"openapi"`
	Servers    []Server    `yaml:"servers,omitempty"`
	Tags       []Tag       `yaml:"tags,omitempty"`
	Paths      Paths       `yaml:"paths"`
}

func NewDocument(document entity.ApiDocument) Document {
	documented := Document{
		OpenApi: openApiVersion,
		Info:    Info{Title: document.Title, Description: document.Description, Version: document.Version},
		Paths:   NewPaths(document.Operations),
	}
	for _, server := range document.Servers {
		documented.Servers = append(documented.Servers, Server{Url: server})
	}
	for _, tag := range document.Tags {
		documented.Tags = append(documented.Tags, NewTag(tag))
	}
	if len(document.SecuritySchemes) > 0 {
		documented.Components = &Components{}
		for _, scheme := range document.SecuritySchemes {
			documented.Components.SecuritySchemes = append(documented.Components.SecuritySchemes, NewSecurityScheme(scheme))
		}
	}
	return documented
}

func EncodeDocument(document Document) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(yamlIndent)
	if err := encoder.Encode(document); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (d *Document) ToEntity() (entity.ApiDocument, error) {
	operations, err := d.Paths.ToEntity()
	if err != nil {
		return entity.ApiDocument{}, err
	}
	document := entity.ApiDocument{
		Title:           d.Info.Title,
		Description:     d.Info.Description,
		Version:         d.Info.Version,
		Servers:         []string{},
		Tags:            []entity.ApiTag{},
		Operations:      operations,
		SecuritySchemes: []entity.ApiSecurityScheme{},
	}
	for _, server := range d.Servers {
		document.Servers = append(document.Servers, server.Url)
	}
	for index := range d.Tags {
		document.Tags = append(document.Tags, d.Tags[index].ToEntity())
	}
	if d.Components != nil {
		for index := range d.Components.SecuritySchemes {
			document.SecuritySchemes = append(document.SecuritySchemes, d.Components.SecuritySchemes[index].ToEntity())
		}
	}
	return document, nil
}

func (d Document) MarshalYAML() (any, error) {
	return mappingOfFields(
		always("openapi", d.OpenApi),
		always("info", d.Info),
		when(len(d.Servers) > 0, "servers", d.Servers),
		when(len(d.Tags) > 0, "tags", d.Tags),
		always("paths", d.Paths),
		when(d.Components != nil, "components", d.Components),
	)
}
