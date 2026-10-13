package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type SnippetLanguageList []SnippetLanguage

func NewSnippetLanguageList(languages []entity.SnippetLanguageSpec) SnippetLanguageList {
	list := make(SnippetLanguageList, 0, len(languages))
	for _, language := range languages {
		list = append(list, NewSnippetLanguage(language))
	}
	return list
}
