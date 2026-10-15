package dto

type MCPWorkspaceEdit struct {
	Id   string `json:"id" description:"Workspace id, from workspace list." used_by:"switch,rename" required_for:"switch,rename" validated_as:"local_id"`
	Name string `json:"name" description:"Workspace name, at most 200 characters." used_by:"create,rename" required_for:"create,rename"`
	Path string `json:"path" description:"Absolute path of the workspace folder." used_by:"create,open" required_for:"create,open"`
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
