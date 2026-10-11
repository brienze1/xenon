package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ScriptSnippet struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Phase string `json:"phase"`
	Text  string `json:"text"`
}

func NewScriptSnippets(snippets []entity.ScriptSnippet) []ScriptSnippet {
	result := make([]ScriptSnippet, 0, len(snippets))
	for _, snippet := range snippets {
		result = append(result, ScriptSnippet{Id: snippet.ID, Title: snippet.Title, Phase: snippet.Phase.String(), Text: snippet.Text})
	}
	return result
}
