package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type WorkspaceSnapshot struct {
	Summary WorkspaceSummary `json:"summary"`
}

func NewWorkspaceSnapshot(workspace entity.RegisteredWorkspace) WorkspaceSnapshot {
	return WorkspaceSnapshot{Summary: NewWorkspaceSummary(workspace)}
}
