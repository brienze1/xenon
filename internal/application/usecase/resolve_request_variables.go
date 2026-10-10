package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

// ResolveRequestVariables resolves copies of the request and of the ancestors whose auth and headers it inherits.
func ResolveRequestVariables(request entity.Request, ancestors []entity.Collection, context entity.VariableContext, dynamic DynamicValue) (entity.Request, []entity.Collection) {
	resolve := func(text string) string {
		return substituteVariables(text, context, dynamic, maxResolutionPasses)
	}
	resolvedAncestors := make([]entity.Collection, 0, len(ancestors))
	for _, ancestor := range ancestors {
		resolvedAncestors = append(resolvedAncestors, ancestor.MapText(resolve))
	}
	return request.MapText(resolve), resolvedAncestors
}
