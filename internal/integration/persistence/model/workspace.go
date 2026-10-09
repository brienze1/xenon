package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
)

type Workspace struct {
	Kind        string   `yaml:"kind"`
	Id          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Order       []string `yaml:"order"`
	Xenon       int      `yaml:"xenon"`
}

func NewWorkspace(workspace entity.Workspace) Workspace {
	return Workspace{
		Xenon:       formatVersion,
		Kind:        itemkind.Workspace.String(),
		Id:          workspace.ID,
		Name:        workspace.Name,
		Description: workspace.Description,
		Order:       nonNil(workspace.Order),
	}
}

func (w *Workspace) ToEntity() entity.Workspace {
	return entity.Workspace{ID: w.Id, Name: w.Name, Description: w.Description, Order: nonNil(w.Order)}
}

func (w *Workspace) Fields() []yamlnode.Field {
	return append(headerFields(itemkind.Workspace.String(), w.Id, w.Name, w.Description),
		yamlnode.StringList("order", nonNil(w.Order)),
		yamlnode.Reserved("variables", yamlnode.EmptyList()),
		yamlnode.Reserved("sync", yamlnode.EmptyMapping()),
	)
}
