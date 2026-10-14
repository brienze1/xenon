package dto

type MCPGitResolution struct {
	Id     string `json:"id"`
	Choice string `json:"choice"`
	Field  string `json:"field"`
}

func (m *MCPGitResolution) ResolveSyncConflict(sessionId string) ResolveSyncConflict {
	return NewResolveSyncConflict(sessionId, m.Id, m.Choice, m.Field)
}
