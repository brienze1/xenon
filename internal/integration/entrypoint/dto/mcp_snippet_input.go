package dto

type MCPSnippetInput struct {
	MCPSendTarget
	Language       string `json:"language"`
	IncludeSecrets bool   `json:"include_secrets"`
}

func (m *MCPSnippetInput) RequestsSecrets() bool {
	return m.IncludeSecrets
}

func (m *MCPSnippetInput) GenerateSnippet(localId, tabId string, request Request) GenerateSnippet {
	return NewGenerateSnippet(localId, tabId, m.Language, request)
}
