package dto

func (m *MCPTreeEdit) EditsRequest() bool {
	return m.Method != nil || m.Url != nil || m.Description != nil || m.Params != nil || m.Headers != nil || m.Body != nil || m.Auth != nil || m.Scripts != nil
}

func (m *MCPTreeEdit) CreateRequest(localId string) CreateRequest {
	return NewCreateRequest(localId, m.ParentId, m.ItemName())
}

// SaveCreatedRequest applies the edits to a request just created, which holds only its id and name.
func (m *MCPTreeEdit) SaveCreatedRequest(localId string, created TreeNode, stored Request) SaveRequest {
	return NewSaveRequest(localId, stored.Revision, NewEditedRequest(NewCreatedRequest(created), m))
}

func (m *MCPTreeEdit) SaveRequest(localId string, stored Request) SaveRequest {
	return NewSaveRequest(localId, stored.Revision, NewEditedRequest(stored, m))
}
