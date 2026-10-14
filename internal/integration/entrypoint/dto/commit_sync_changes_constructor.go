package dto

func NewCommitSyncChanges(localId, message string, itemIds []string) CommitSyncChanges {
	return CommitSyncChanges{LocalId: localId, Message: message, ItemIDs: itemIds}
}
