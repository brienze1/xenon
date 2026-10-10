package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

const maxResolutionPasses = 19

type DynamicValue func(name string) (string, bool)

// ResolveText replaces references in up to 19 passes (Postman's limit); unknown names stay literal.
func ResolveText(text string, context entity.VariableContext, dynamic DynamicValue) string {
	return substituteVariables(text, context, dynamic, maxResolutionPasses)
}

func substituteVariables(text string, context entity.VariableContext, dynamic DynamicValue, passesLeft int) string {
	if passesLeft == 0 {
		return text
	}
	next := variableReference.ReplaceAllStringFunc(text, func(match string) string {
		if value, found := lookupVariable(match[2:len(match)-2], context, dynamic); found {
			return value
		}
		return match
	})
	if next == text {
		return text
	}
	return substituteVariables(next, context, dynamic, passesLeft-1)
}

func lookupVariable(name string, context entity.VariableContext, dynamic DynamicValue) (string, bool) {
	if variable, found := context.Lookup(name); found {
		return variable.Value, variable.IsResolvable()
	}
	if dynamic == nil {
		return "", false
	}
	return dynamic(name)
}
