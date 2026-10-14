package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPWorkspace struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func NewMCPWorkspace(workspace entity.RegisteredWorkspace) MCPWorkspace {
	return MCPWorkspace{Id: workspace.LocalID, Name: workspace.DisplayNameCache, Path: workspace.Path}
}
