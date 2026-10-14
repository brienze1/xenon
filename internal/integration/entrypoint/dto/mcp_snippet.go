package dto

type MCPSnippet struct {
	Code    string   `json:"code"`
	Notices []string `json:"notices,omitempty"`
}

func NewMCPSnippet(snippet Snippet) MCPSnippet {
	return MCPSnippet{Code: snippet.Text, Notices: snippet.Notices}
}
