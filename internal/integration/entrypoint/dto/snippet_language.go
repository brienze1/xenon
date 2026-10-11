package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SnippetLanguage struct {
	Id      string          `json:"id"`
	Label   string          `json:"label"`
	Options []SnippetOption `json:"options"`
}

func NewSnippetLanguage(language entity.SnippetLanguageSpec) SnippetLanguage {
	options := make([]SnippetOption, 0, len(language.Options))
	for _, option := range language.Options {
		options = append(options, NewSnippetOption(option))
	}
	return SnippetLanguage{Id: language.Language.String(), Label: language.Label, Options: options}
}
