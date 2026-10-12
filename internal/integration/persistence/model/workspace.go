package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type Workspace struct {
	Postman     map[string]any `yaml:"postman"`
	Sync        WorkspaceSync  `yaml:"sync"`
	Kind        string         `yaml:"kind"`
	Id          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Order       []string       `yaml:"order"`
	Variables   []VariableRow  `yaml:"variables"`
	Xenon       int            `yaml:"xenon"`
}

func NewWorkspace(workspace entity.Workspace) Workspace {
	return Workspace{
		Xenon:       formatVersion,
		Kind:        itemkind.Workspace.String(),
		Id:          workspace.ID,
		Name:        workspace.Name,
		Description: workspace.Description,
		Order:       nonNil(workspace.Order),
		Variables:   newVariableRows(workspace.Variables),
		Postman:     workspace.Postman,
	}
}

func (w *Workspace) ToEntity() entity.Workspace {
	return entity.Workspace{ID: w.Id, Name: w.Name, Description: w.Description, Order: nonNil(w.Order), Variables: variableRowsToEntity(w.Variables), Postman: sourceFieldsToEntity(w.Postman), Sync: w.Sync.ToEntity()}
}

func (w *Workspace) Fields() []yamlnode.Field {
	fields := append(headerFields(itemkind.Workspace.String(), w.Id, w.Name, w.Description),
		yamlnode.StringList("order", nonNil(w.Order)),
		variablesField(w.Variables),
		yamlnode.Reserved("sync", yamlnode.EmptyMapping()),
	)
	return append(fields, optionalSourceFields(postmanKey, w.Postman)...)
}
