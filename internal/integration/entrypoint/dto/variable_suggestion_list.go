package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type VariableSuggestionList []VariableSuggestion

func NewVariableSuggestionList(suggestions []entity.VariableSuggestion) VariableSuggestionList {
	list := make(VariableSuggestionList, 0, len(suggestions))
	for _, suggestion := range suggestions {
		list = append(list, NewVariableSuggestion(suggestion))
	}
	return list
}

func (v VariableSuggestionList) LogValue() any {
	return map[string]any{"suggestions": len(v)}
}
