package dto

type MCPGitCommit struct {
	Message string   `json:"message" description:"Commit message." required_for:"commit"`
	ItemIds []string `json:"item_ids" description:"Ids of the changed items to commit; omitted commits every change."`
}

func (m *MCPGitCommit) HasItemIds() bool {
	return len(m.ItemIds) > 0
}

// CommitSyncChanges falls back to changedIds when the arguments name no items.
func (m *MCPGitCommit) CommitSyncChanges(localId string, changedIds []string) CommitSyncChanges {
	itemIds := m.ItemIds
	if !m.HasItemIds() {
		itemIds = changedIds
	}
	return NewCommitSyncChanges(localId, m.Message, itemIds)
}
