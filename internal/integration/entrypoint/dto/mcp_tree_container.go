package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPTreeContainer struct {
	Id          string `json:"id"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
	Children    int    `json:"children"`
}

func NewMCPTreeContainer(entry entity.TreeEntry, settings ItemSettings) MCPTreeContainer {
	return MCPTreeContainer{
		Id:          entry.ID(),
		Kind:        entry.Kind().String(),
		Name:        entry.Name(),
		Description: settings.Description,
		Path:        entry.Path,
		Children:    entry.Children(),
	}
}
