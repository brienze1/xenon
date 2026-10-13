package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type AuthUnsupported struct {
	Data   any    `yaml:"data"`
	Source string `yaml:"source"`
	Type   string `yaml:"type"`
}

func NewAuthUnsupported(unsupported entity.AuthUnsupported) AuthUnsupported {
	return AuthUnsupported{Source: unsupported.Origin.String(), Type: unsupported.Type, Data: unsupported.Data}
}

func (a *AuthUnsupported) ToEntity() entity.AuthUnsupported {
	return entity.AuthUnsupported{Origin: importorigin.GetImportOrigin(a.Source), Type: a.Type, Data: a.Data}
}

func (a *AuthUnsupported) Node() *yaml.Node {
	return yamlnode.MappingNode(
		yamlnode.String("source", a.Source),
		yamlnode.String("type", a.Type),
		yamlnode.Field{Key: "data", Value: yamlnode.Any(a.Data)},
	)
}
