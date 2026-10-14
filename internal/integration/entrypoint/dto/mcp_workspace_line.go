package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPWorkspaceLine struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active,omitempty"`
}

func NewMCPWorkspaceLine(workspace entity.RegisteredWorkspace, active bool) MCPWorkspaceLine {
	return MCPWorkspaceLine{Id: workspace.LocalID, Name: workspace.DisplayNameCache, Active: active}
}

func NewMCPWorkspaceList(registry entity.AppRegistry) MCPPage[MCPWorkspaceLine] {
	lines := make([]MCPWorkspaceLine, 0, len(registry.Workspaces))
	for _, workspace := range registry.Workspaces {
		lines = append(lines, NewMCPWorkspaceLine(workspace, registry.IsActive(workspace.LocalID)))
	}
	return NewMCPList(lines)
}
