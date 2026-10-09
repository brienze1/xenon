package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type WorkspaceSummary struct {
	LocalId string `json:"local_id"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Missing bool   `json:"missing"`
}

func NewWorkspaceSummary(workspace entity.RegisteredWorkspace) WorkspaceSummary {
	return WorkspaceSummary{
		LocalId: workspace.LocalID,
		Name:    workspace.DisplayNameCache,
		Path:    workspace.Path,
		Missing: workspace.Missing,
	}
}
