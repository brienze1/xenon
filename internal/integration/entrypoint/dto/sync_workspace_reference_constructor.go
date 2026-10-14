package dto

func NewSyncWorkspaceReference(localId string) SyncWorkspaceReference {
	return SyncWorkspaceReference{LocalId: localId}
}
