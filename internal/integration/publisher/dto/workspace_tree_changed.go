package dto

const WorkspaceTreeChangedEvent = "workspace:tree-changed"

type WorkspaceTreeChanged struct {
	LocalId string `json:"local_id"`
}

func NewWorkspaceTreeChanged(localId string) WorkspaceTreeChanged {
	return WorkspaceTreeChanged{LocalId: localId}
}
