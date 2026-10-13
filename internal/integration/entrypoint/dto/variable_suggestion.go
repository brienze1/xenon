package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableSuggestion struct {
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
	Secret      bool   `json:"secret"`
	Dynamic     bool   `json:"dynamic"`
}

func NewVariableSuggestion(suggestion entity.VariableSuggestion) VariableSuggestion {
	return VariableSuggestion{
		Name:        suggestion.Name,
		Scope:       suggestion.Scope.String(),
		Description: suggestion.Description,
		Secret:      suggestion.Secret,
		Dynamic:     suggestion.Dynamic,
	}
}
