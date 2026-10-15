package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

const environmentOrderKey = "order"

type Workspace struct {
	Postman          map[string]any `yaml:"postman"`
	Sync             WorkspaceSync  `yaml:"sync"`
	Kind             string         `yaml:"kind"`
	Id               string         `yaml:"id"`
	Name             string         `yaml:"name"`
	Description      string         `yaml:"description"`
	Order            []string       `yaml:"order"`
	EnvironmentOrder []string       `yaml:"-"`
	Variables        []VariableRow  `yaml:"variables"`
	Environments     yaml.Node      `yaml:"environments"`
	Xenon            int            `yaml:"xenon"`
}

func NewWorkspace(workspace entity.Workspace) Workspace {
	return Workspace{
		Xenon:            formatVersion,
		Kind:             itemkind.Workspace.String(),
		Id:               workspace.ID,
		Name:             workspace.Name,
		Description:      workspace.Description,
		Order:            nonNil(workspace.Order),
		EnvironmentOrder: workspace.EnvironmentOrder,
		Variables:        newVariableRows(workspace.Variables),
		Postman:          workspace.Postman,
	}
}

func (w *Workspace) ToEntity() entity.Workspace {
	return entity.Workspace{
		ID:               w.Id,
		Name:             w.Name,
		Description:      w.Description,
		Order:            nonNil(w.Order),
		EnvironmentOrder: w.storedEnvironmentOrder(),
		Variables:        variableRowsToEntity(w.Variables),
		Postman:          sourceFieldsToEntity(w.Postman),
		Sync:             w.Sync.ToEntity(),
	}
}

func (w *Workspace) Fields() []yamlnode.Field {
	fields := append(headerFields(itemkind.Workspace.String(), w.Id, w.Name, w.Description), yamlnode.StringList("order", nonNil(w.Order)))
	fields = append(fields, w.environmentsFields()...)
	fields = append(fields, variablesField(w.Variables), yamlnode.Reserved("sync", yamlnode.EmptyMapping()))
	return append(fields, optionalSourceFields(postmanKey, w.Postman)...)
}

func (w *Workspace) environmentsFields() []yamlnode.Field {
	if w.EnvironmentOrder == nil {
		return []yamlnode.Field{}
	}
	return []yamlnode.Field{yamlnode.Object("environments", yamlnode.StringList(environmentOrderKey, w.EnvironmentOrder))}
}

func (w *Workspace) storedEnvironmentOrder() []string {
	order := mappingValue(&w.Environments, environmentOrderKey)
	if order == nil || order.Kind != yaml.SequenceNode {
		return nil
	}
	ids := make([]string, 0, len(order.Content))
	for _, item := range order.Content {
		if item.Kind != yaml.ScalarNode {
			return nil
		}
		ids = append(ids, item.Value)
	}
	return ids
}
