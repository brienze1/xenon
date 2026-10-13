package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ApiDocument struct {
	Title           string              `json:"title"`
	Description     string              `json:"description"`
	Version         string              `json:"version"`
	Servers         []string            `json:"servers"`
	Tags            []ApiTag            `json:"tags"`
	TagGroups       []ApiTagGroup       `json:"tag_groups"`
	Operations      []ApiOperation      `json:"operations"`
	SecuritySchemes []ApiSecurityScheme `json:"security_schemes"`
}

func NewApiDocument(document entity.ApiDocument) ApiDocument {
	documented := ApiDocument{
		Title:           document.Title,
		Description:     document.Description,
		Version:         document.Version,
		Servers:         append([]string{}, document.Servers...),
		Tags:            make([]ApiTag, 0, len(document.Tags)),
		TagGroups:       make([]ApiTagGroup, 0, len(document.TagGroups)),
		Operations:      make([]ApiOperation, 0, len(document.Operations)),
		SecuritySchemes: make([]ApiSecurityScheme, 0, len(document.SecuritySchemes)),
	}
	for _, tag := range document.Tags {
		documented.Tags = append(documented.Tags, NewApiTag(tag))
	}
	for _, group := range document.TagGroups {
		documented.TagGroups = append(documented.TagGroups, NewApiTagGroup(group))
	}
	for _, operation := range document.Operations {
		documented.Operations = append(documented.Operations, NewApiOperation(operation, document.ServerOf(operation)))
	}
	for _, scheme := range document.SecuritySchemes {
		documented.SecuritySchemes = append(documented.SecuritySchemes, NewApiSecurityScheme(scheme))
	}
	return documented
}
