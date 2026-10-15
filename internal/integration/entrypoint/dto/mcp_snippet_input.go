package dto

type MCPSnippetInput struct {
	MCPSendTarget
	Language       string `json:"language" description:"Snippet language." required_for:"generate" enum:"curl,js_fetch,go_native,python_requests"`
	IncludeSecrets bool   `json:"include_secrets" description:"Refused over MCP: only a human can include secrets."`
}

func (m *MCPSnippetInput) RequestsSecrets() bool {
	return m.IncludeSecrets
}

func (m *MCPSnippetInput) GenerateSnippet(localId, tabId string, request Request) GenerateSnippet {
	return NewGenerateSnippet(localId, tabId, m.Language, request)
}
