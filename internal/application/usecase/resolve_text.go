package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

func ResolveText(text string, context entity.VariableContext, dynamic entity.DynamicValue) string {
	return context.Resolve(text, dynamic)
}
