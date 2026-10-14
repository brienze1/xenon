package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPTreeEdited struct {
	Id     string `json:"id"`
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
}

func NewMCPTreeEdited(entry entity.TreeEntry, method string) MCPTreeEdited {
	return MCPTreeEdited{Id: entry.ID(), Kind: entry.Kind().String(), Name: entry.Name(), Method: method, Path: entry.Path}
}
