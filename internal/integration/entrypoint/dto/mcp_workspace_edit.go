package dto

type MCPWorkspaceEdit struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (m *MCPWorkspaceEdit) CreateWorkspace() CreateWorkspace {
	return NewCreateWorkspace(m.Name, m.Path)
}

func (m *MCPWorkspaceEdit) OpenWorkspace() OpenWorkspace {
	return NewOpenWorkspace(m.Path)
}

func (m *MCPWorkspaceEdit) RenameWorkspace() RenameWorkspace {
	return NewRenameWorkspace(m.Id, m.Name)
}

func (m *MCPWorkspaceEdit) ActivateWorkspace() ActivateWorkspace {
	return NewActivateWorkspace(m.Id)
}
