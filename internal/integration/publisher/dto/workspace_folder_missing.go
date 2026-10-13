package dto

const WorkspaceFolderMissingEvent = "workspace:folder-missing"

type WorkspaceFolderMissing struct {
	LocalId string `json:"local_id"`
}

func NewWorkspaceFolderMissing(localId string) WorkspaceFolderMissing {
	return WorkspaceFolderMissing{LocalId: localId}
}
