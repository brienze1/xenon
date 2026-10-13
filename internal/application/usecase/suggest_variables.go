package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const maxSuggestions = 50

// SuggestVariables offers the names starting with the prefix (ignoring case): user variables by name, strongest scope first, then the dynamic variables they do not shadow, at most 50.
func SuggestVariables(context entity.VariableContext, dynamicNames []string, prefix string) []entity.VariableSuggestion {
	suggestions := make([]entity.VariableSuggestion, 0)
	for _, variable := range context.Variables() {
		if hasPrefixIgnoringCase(variable.Name, prefix) {
			suggestions = append(suggestions, entity.NewVariableSuggestion(variable))
		}
	}
	for _, name := range dynamicNames {
		if _, shadowed := context.Lookup(name); !shadowed && hasPrefixIgnoringCase(name, prefix) {
			suggestions = append(suggestions, entity.NewDynamicVariableSuggestion(name))
		}
	}
	return suggestions[:min(len(suggestions), maxSuggestions)]
}

func hasPrefixIgnoringCase(name, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix))
}
