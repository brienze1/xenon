package dto

type MCPSnippetLanguage struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func NewMCPSnippetLanguage(language SnippetLanguage) MCPSnippetLanguage {
	return MCPSnippetLanguage{Id: language.Id, Name: language.Label}
}

func NewMCPSnippetLanguages(languages SnippetLanguageList) []MCPSnippetLanguage {
	lines := make([]MCPSnippetLanguage, 0, len(languages))
	for _, language := range languages {
		lines = append(lines, NewMCPSnippetLanguage(language))
	}
	return lines
}
