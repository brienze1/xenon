package dto

type MCPGitResolve struct {
	Items []MCPGitResolution `json:"items"`
}

func (m *MCPGitResolve) ResolveSyncConflicts(sessionId string) []ResolveSyncConflict {
	conflicts := make([]ResolveSyncConflict, 0, len(m.Items))
	for _, item := range m.Items {
		conflicts = append(conflicts, item.ResolveSyncConflict(sessionId))
	}
	return conflicts
}
