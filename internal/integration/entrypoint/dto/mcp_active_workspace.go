package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPActiveWorkspace struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	EnvironmentId string `json:"environment_id,omitempty"`
}

func NewMCPActiveWorkspace(workspace entity.RegisteredWorkspace, environments entity.EnvironmentList) MCPActiveWorkspace {
	return MCPActiveWorkspace{Id: workspace.LocalID, Name: workspace.DisplayNameCache, Path: workspace.Path, EnvironmentId: environments.ActiveID}
}
