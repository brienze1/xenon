package entity

type WorkspaceSnapshot struct {
	Workspace RegisteredWorkspace
	Tree      Tree
	UIState   UIState
}

func (w WorkspaceSnapshot) LogValue() any {
	return map[string]any{"workspace": w.Workspace, "tree": w.Tree, "uiState": w.UIState.LogValue()}
}
