package dto

func NewGenerateSnippet(localId, tabId, language string, request Request) GenerateSnippet {
	return GenerateSnippet{LocalId: localId, TabId: tabId, Language: language, Request: request}
}
