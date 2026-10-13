package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type SnippetLanguageSpec struct {
	Language enums.SnippetLanguage
	Label    string
	Options  []SnippetOptionSpec
}
