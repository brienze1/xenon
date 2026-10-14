package dto

type MCPScriptSnippet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func NewMCPScriptSnippet(snippet ScriptSnippet) MCPScriptSnippet {
	return MCPScriptSnippet{Id: snippet.Id, Name: snippet.Title, Code: snippet.Text}
}

func NewMCPScriptSnippets(snippets []ScriptSnippet) []MCPScriptSnippet {
	lines := make([]MCPScriptSnippet, 0, len(snippets))
	for _, snippet := range snippets {
		lines = append(lines, NewMCPScriptSnippet(snippet))
	}
	return lines
}
