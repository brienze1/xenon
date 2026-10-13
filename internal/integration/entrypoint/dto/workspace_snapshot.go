package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type WorkspaceSnapshot struct {
	Summary WorkspaceSummary `json:"summary"`
	Tree    TreeNodeList     `json:"tree"`
	UIState UIState          `json:"ui_state"`
}

func NewWorkspaceSnapshot(snapshot entity.WorkspaceSnapshot) WorkspaceSnapshot {
	return WorkspaceSnapshot{
		Summary: NewWorkspaceSummary(snapshot.Workspace),
		Tree:    NewTreeNodeList(snapshot.Tree),
		UIState: NewUIState(snapshot.Workspace.LocalID, snapshot.UIState),
	}
}

func (w WorkspaceSnapshot) LogValue() any {
	return map[string]any{"summary": w.Summary, "tree": w.Tree.LogValue(), "ui_state": w.UIState.LogValue()}
}
