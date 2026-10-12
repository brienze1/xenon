package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SecurityScheme struct {
	Name        string `yaml:"-"`
	Type        string `yaml:"type"`
	Description string `yaml:"description,omitempty"`
	Scheme      string `yaml:"scheme,omitempty"`
	ParamName   string `yaml:"name,omitempty"`
	In          string `yaml:"in,omitempty"`
	Flows       Flows  `yaml:"flows,omitempty"`
}

func NewSecurityScheme(scheme entity.ApiSecurityScheme) SecurityScheme {
	documented := SecurityScheme{
		Name:        scheme.Name,
		Type:        scheme.Type,
		Description: scheme.Description,
		Scheme:      scheme.Scheme,
		ParamName:   scheme.ParamName,
		In:          scheme.In,
	}
	if scheme.Flow != "" {
		documented.Flows = Flows{{Name: scheme.Flow, AuthorizationUrl: scheme.AuthorizationURL, TokenUrl: scheme.TokenURL, Scopes: append(Scopes{}, scheme.Scopes...)}}
	}
	return documented
}

func (s *SecurityScheme) ToEntity() entity.ApiSecurityScheme {
	scheme := entity.ApiSecurityScheme{
		Name:        s.Name,
		Type:        s.Type,
		Description: s.Description,
		Scheme:      s.Scheme,
		ParamName:   s.ParamName,
		In:          s.In,
		Scopes:      []string{},
	}
	if len(s.Flows) > 0 {
		flow := s.Flows[0]
		scheme.Flow, scheme.AuthorizationURL, scheme.TokenURL, scheme.Scopes = flow.Name, flow.AuthorizationUrl, flow.TokenUrl, append([]string{}, flow.Scopes...)
	}
	return scheme
}

func (s SecurityScheme) MarshalYAML() (any, error) {
	return mappingOfFields(
		always("type", s.Type),
		text("description", s.Description),
		text("scheme", s.Scheme),
		text("name", s.ParamName),
		text("in", s.In),
		when(len(s.Flows) > 0, "flows", s.Flows),
	)
}
