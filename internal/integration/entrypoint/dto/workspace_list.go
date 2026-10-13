package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type WorkspaceList struct {
	ActiveLocalId string             `json:"active_local_id"`
	Workspaces    []WorkspaceSummary `json:"workspaces"`
}

func NewWorkspaceList(registry entity.AppRegistry) WorkspaceList {
	list := WorkspaceList{ActiveLocalId: registry.ActiveLocalID, Workspaces: make([]WorkspaceSummary, 0, len(registry.Workspaces))}
	for _, workspace := range registry.Workspaces {
		list.Workspaces = append(list.Workspaces, NewWorkspaceSummary(workspace))
	}
	return list
}
