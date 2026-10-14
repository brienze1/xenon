package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPGitPull struct {
	Session *MCPGitSession `json:"session,omitempty"`
	MCPGitStatus
}

func NewMCPGitPull(pulled entity.PullResult) MCPGitPull {
	converted := MCPGitPull{MCPGitStatus: NewMCPGitStatus(pulled.Status)}
	if pulled.Session != nil {
		session := NewMCPGitSession(pulled.Session)
		converted.Session = &session
	}
	return converted
}
