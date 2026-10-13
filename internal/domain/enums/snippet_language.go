package enums

import "strings"

type SnippetLanguage string

func (l SnippetLanguage) Equals(s string) bool {
	return strings.EqualFold(string(l), s)
}

func (l SnippetLanguage) String() string {
	return string(l)
}
