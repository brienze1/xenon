package dto

const WorkspacesChangedEvent = "workspaces:changed"

type WorkspacesChanged struct {
	ActiveLocalId string `json:"active_local_id"`
}

func NewWorkspacesChanged(activeLocalId string) WorkspacesChanged {
	return WorkspacesChanged{ActiveLocalId: activeLocalId}
}
