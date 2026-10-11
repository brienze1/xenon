package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

type SnippetOptions struct {
	Options  map[string]string `json:"options"`
	Language string            `json:"language" validate:"oneof=curl js_fetch go_native python_requests"`
}

func NewSnippetOptions(options entity.SnippetOptions) SnippetOptions {
	return SnippetOptions{Language: options.Language.String(), Options: options.Options}
}

func (s *SnippetOptions) ToEntity() entity.SnippetOptions {
	return entity.SnippetOptions{Language: snippetlanguage.GetSnippetLanguage(s.Language), Options: s.Options}
}
