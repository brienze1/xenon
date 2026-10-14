package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPGitSession struct {
	Id    string           `json:"id,omitempty"`
	Items []MCPGitConflict `json:"items"`
}

func NewMCPGitSession(session *entity.SyncSession) MCPGitSession {
	if session == nil {
		return MCPGitSession{Items: []MCPGitConflict{}}
	}
	items := make([]MCPGitConflict, 0, len(session.Conflicts))
	for _, conflict := range session.Conflicts {
		items = append(items, NewMCPGitConflict(conflict))
	}
	return MCPGitSession{Id: session.ID, Items: items}
}
