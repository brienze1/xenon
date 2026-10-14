package dto

type MCPDocsInput struct {
	CollectionId   string `json:"collection_id"`
	Yaml           string `json:"yaml"`
	ConfirmDeletes bool   `json:"confirm_deletes"`
}

func (m *MCPDocsInput) DocumentationReference(localId string) TreeItemReference {
	return NewTreeItemReference(localId, m.CollectionId)
}

func (m *MCPDocsInput) PlanInput(localId string) DocumentationInput {
	return NewDocumentationInput(localId, m.CollectionId, m.Yaml, false, false)
}

func (m *MCPDocsInput) ApplyInput(localId string) DocumentationInput {
	return NewDocumentationInput(localId, m.CollectionId, m.Yaml, m.ConfirmDeletes, !m.ConfirmDeletes)
}

func (m *MCPDocsInput) CollectionItemId() string {
	return m.CollectionId
}
