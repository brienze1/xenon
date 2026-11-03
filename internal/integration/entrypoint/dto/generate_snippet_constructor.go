package dto

func NewGenerateSnippet(localId, tabId, language string, request Request) GenerateSnippet {
	return GenerateSnippet{LocalId: localId, TabId: tabId, Language: language, Request: request}
}

func (g GenerateSnippet) InEnvironment(environmentId string) GenerateSnippet {
	g.EnvironmentId = environmentId
	return g
}
