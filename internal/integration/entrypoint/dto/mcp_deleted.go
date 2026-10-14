package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPDeleted struct {
	Deleted string `json:"deleted"`
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Path    string `json:"path,omitempty"`
}

func NewMCPDeleted(deleted string, target entity.AIActivityTarget) MCPDeleted {
	return MCPDeleted{Deleted: deleted, Id: target.ID, Name: target.Name, Path: target.Path}
}
