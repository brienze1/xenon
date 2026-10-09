package entity

type WorkspaceSnapshot struct {
	Workspace RegisteredWorkspace
	Tree      Tree
	UIState   UIState
}
