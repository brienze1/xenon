package usecase

import "github.com/brienze1/xenon/internal/domain/entity"

// InspectRequestVariables reports every reference of the request and of what it inherits; offsets count UTF-16 units.
func InspectRequestVariables(request entity.Request, ancestors []entity.Collection, context entity.VariableContext, dynamicNames []string) entity.VariableInspection {
	inspection := entity.NewVariableInspection(context)
	for _, field := range append(request.TextFields(), request.InheritedFields(ancestors)...) {
		for _, reference := range field.References() {
			resolved := context.Inspect(reference.WithUtf16Offsets(field.Text), dynamicNames)
			inspection = inspection.WithReference(resolved).WithNestedUsages(resolved, field.Name, context, dynamicNames)
		}
	}
	return inspection
}
