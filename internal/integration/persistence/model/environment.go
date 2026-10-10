package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Environment struct {
	Postman           map[string]any `yaml:"postman"`
	Bruno             map[string]any `yaml:"bruno"`
	Kind              string         `yaml:"kind"`
	Id                string         `yaml:"id"`
	Name              string         `yaml:"name"`
	Description       string         `yaml:"description"`
	Color             string         `yaml:"color"`
	Variables         []VariableRow  `yaml:"variables"`
	Xenon             int            `yaml:"xenon"`
	ConfirmBeforeSend bool           `yaml:"confirmBeforeSend"`
}

func NewEnvironment(environment entity.Environment) Environment {
	return Environment{
		Xenon:             formatVersion,
		Kind:              itemkind.Environment.String(),
		Id:                environment.ID,
		Name:              environment.Name,
		Description:       environment.Description,
		Color:             environment.Color.String(),
		ConfirmBeforeSend: environment.ConfirmBeforeSend,
		Variables:         newVariableRows(environment.Variables),
		Postman:           environment.Postman,
		Bruno:             environment.Bruno,
	}
}

func (e *Environment) ToEntity(fileName string) entity.Environment {
	color := envcolor.GetEnvColor(e.Color)
	if color == envcolor.Unknown {
		color = envcolor.None
	}
	return entity.Environment{
		ID:                e.Id,
		Name:              e.Name,
		Description:       e.Description,
		Color:             color,
		ConfirmBeforeSend: e.ConfirmBeforeSend,
		Variables:         variableRowsToEntity(e.Variables),
		FileName:          fileName,
		Postman:           sourceFieldsToEntity(e.Postman),
		Bruno:             sourceFieldsToEntity(e.Bruno),
	}
}

func (e *Environment) Fields() []yamlnode.Field {
	fields := append(headerFields(itemkind.Environment.String(), e.Id, e.Name, e.Description),
		yamlnode.String("color", e.Color),
		yamlnode.Bool("confirmBeforeSend", e.ConfirmBeforeSend),
		variablesField(e.Variables),
	)
	fields = append(fields, optionalSourceFields(postmanKey, e.Postman)...)
	return append(fields, optionalSourceFields(brunoKey, e.Bruno)...)
}
