package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/snippetphase"
)

type ScriptSnippet struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Phase string `json:"phase"`
	Text  string `json:"text"`
}

func (s ScriptSnippet) ToEntity() entity.ScriptSnippet {
	return entity.ScriptSnippet{ID: s.Id, Title: s.Title, Phase: snippetphase.GetSnippetPhase(s.Phase), Text: s.Text}
}
