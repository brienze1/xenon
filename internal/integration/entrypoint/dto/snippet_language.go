package dto

import (
	"slices"
	"strconv"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	booleanSnippetOption  = "boolean"
	positiveSnippetOption = "positiveInteger"
)

type SnippetOption struct {
	Default any      `json:"default"`
	Id      string   `json:"id"`
	Label   string   `json:"label"`
	Type    string   `json:"type"`
	Values  []string `json:"values,omitempty"`
}

type SnippetLanguage struct {
	Id      string          `json:"id"`
	Label   string          `json:"label"`
	Options []SnippetOption `json:"options"`
}

type SnippetLanguageList []SnippetLanguage

func NewSnippetLanguageList(languages []entity.SnippetLanguageSpec) SnippetLanguageList {
	list := make(SnippetLanguageList, 0, len(languages))
	for _, language := range languages {
		options := make([]SnippetOption, 0, len(language.Options))
		for _, option := range language.Options {
			options = append(options, SnippetOption{Id: option.ID, Label: option.Label, Type: option.Type, Default: typedDefault(option), Values: slices.Clone(option.Values)})
		}
		list = append(list, SnippetLanguage{Id: language.Language.String(), Label: language.Label, Options: options})
	}
	return list
}

func typedDefault(option entity.SnippetOptionSpec) any {
	switch option.Type {
	case booleanSnippetOption:
		return option.Default == "true"
	case positiveSnippetOption:
		number, _ := strconv.ParseFloat(option.Default, 64)
		return number
	default:
		return option.Default
	}
}
