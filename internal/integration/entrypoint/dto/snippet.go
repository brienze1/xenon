package dto

import (
	"slices"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Snippet struct {
	Text    string   `json:"text"`
	Notices []string `json:"notices"`
}

func NewSnippet(snippet entity.Snippet) Snippet {
	notices := slices.Clone(snippet.Notices)
	if notices == nil {
		notices = []string{}
	}
	return Snippet{Text: snippet.Text, Notices: notices}
}

func (s Snippet) LogValue() any {
	return map[string]any{"length": len(s.Text), "notices": s.Notices}
}
