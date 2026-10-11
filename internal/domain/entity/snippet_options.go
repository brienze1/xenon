package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type SnippetOptions struct {
	Options  map[string]string
	Language enums.SnippetLanguage
}
