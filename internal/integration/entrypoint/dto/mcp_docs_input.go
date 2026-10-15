package dto

type MCPDocsInput struct {
	CollectionId   string `json:"collection_id" description:"Collection id." required_for:"get,plan,apply" validated_as:"id"`
	Yaml           string `json:"yaml" description:"The edited OpenAPI document as YAML." used_by:"plan,apply"`
	ConfirmDeletes bool   `json:"confirm_deletes" description:"Delete the requests the document no longer lists." schema_for:"docs_edit"`
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
