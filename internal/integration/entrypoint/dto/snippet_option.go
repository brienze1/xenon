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
	Default     any      `json:"default"`
	Description string   `json:"description"`
	Id          string   `json:"id"`
	Label       string   `json:"label"`
	Type        string   `json:"type"`
	Values      []string `json:"values,omitempty"`
}

func NewSnippetOption(option entity.SnippetOptionSpec) SnippetOption {
	return SnippetOption{Id: option.ID, Label: option.Label, Description: option.Description, Type: option.Type, Default: typedDefault(option), Values: slices.Clone(option.Values)}
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
