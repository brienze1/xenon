package usecase

import (
	"slices"
	"unicode/utf16"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/referencestatus"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

// InspectRequestVariables reports every reference of the request and of what it inherits, without generating dynamic values; offsets count UTF-16 units, secrets are masked.
func InspectRequestVariables(request entity.Request, ancestors []entity.Collection, context entity.VariableContext, dynamicNames []string) entity.VariableInspection {
	inspection := entity.NewVariableInspection(context)
	for _, field := range append(request.TextFields(), inheritedFields(request, ancestors)...) {
		for _, reference := range TokenizeReferences(field.Name, field.Text) {
			resolved := inspectReference(withUtf16Offsets(field.Text, reference), context, dynamicNames)
			inspection = withNestedUsages(inspection.WithReference(resolved), resolved, field.Name, context, dynamicNames)
		}
	}
	return inspection
}

func inheritedFields(request entity.Request, ancestors []entity.Collection) []entity.TextField {
	authOwner := inheritedAuthOwner(request, ancestors)
	fields := make([]entity.TextField, 0)
	for index, ancestor := range ancestors {
		fields = append(fields, ancestor.InheritedFields(index == authOwner, inheritedHeaders(request, ancestors, index))...)
	}
	return fields
}

func inheritedAuthOwner(request entity.Request, ancestors []entity.Collection) int {
	owner := -1
	if !request.Auth.IsInherited() {
		return owner
	}
	for index, ancestor := range ancestors {
		if !ancestor.Auth.IsInherited() {
			owner = index
		}
	}
	return owner
}

func inheritedHeaders(request entity.Request, ancestors []entity.Collection, position int) []int {
	overriding := slices.Clone(request.Headers)
	for _, descendant := range ancestors[position+1:] {
		overriding = append(overriding, descendant.Headers...)
	}
	indexes := make([]int, 0)
	for index, header := range ancestors[position].Headers {
		if header.IsActive() && !overridden(header, overriding) {
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func overridden(header entity.KeyValue, overriding []entity.KeyValue) bool {
	return slices.ContainsFunc(overriding, func(other entity.KeyValue) bool {
		return other.IsActive() && other.HasKey(header.Key)
	})
}

func withUtf16Offsets(text string, reference entity.Reference) entity.Reference {
	start := utf16Length(text[:reference.Start])
	reference.End = start + utf16Length(text[reference.Start:reference.End])
	reference.Start = start
	return reference
}

func utf16Length(text string) int {
	return len(utf16.Encode([]rune(text)))
}

func inspectReference(reference entity.Reference, context entity.VariableContext, dynamicNames []string) entity.ResolvedReference {
	variable, found := context.Lookup(reference.Name)
	switch {
	case found && variable.Missing:
		return entity.ResolvedReference{Reference: reference, Status: referencestatus.MissingSecret, Scope: variable.Scope, Secret: true}
	case found && variable.Locked:
		return entity.ResolvedReference{Reference: reference, Status: referencestatus.LockedSecret, Scope: variable.Scope, Secret: true}
	case found:
		return resolvedReference(reference, variable, context)
	case slices.Contains(dynamicNames, reference.Name):
		return entity.ResolvedReference{Reference: reference, Status: referencestatus.Dynamic, Scope: variablescope.Dynamic}
	default:
		return entity.ResolvedReference{Reference: reference, Status: referencestatus.Unresolved, Scope: variablescope.None}
	}
}

func resolvedReference(reference entity.Reference, variable entity.ContextVariable, context entity.VariableContext) entity.ResolvedReference {
	resolved := substituteVariables(reference.Placeholder(), context, nil, maxResolutionPasses)
	if substituteVariables(resolved, context, nil, 1) != resolved {
		return entity.ResolvedReference{Reference: reference, Status: referencestatus.Unresolved, Scope: variable.Scope, Secret: variable.Secret}
	}
	return entity.ResolvedReference{Reference: reference, Status: referencestatus.Resolved, Scope: variable.Scope, Secret: variable.Secret, Value: context.Masked(resolved)}
}

func withNestedUsages(inspection entity.VariableInspection, resolved entity.ResolvedReference, field string, context entity.VariableContext, dynamicNames []string) entity.VariableInspection {
	visited := map[string]bool{}
	pending := []entity.ResolvedReference{resolved}
	for len(pending) > 0 {
		current := pending[0]
		pending = pending[1:]
		if visited[current.Reference.Name] {
			continue
		}
		visited[current.Reference.Name] = true
		inspection = inspection.WithUsage(current, field)
		variable, found := context.Lookup(current.Reference.Name)
		if !found {
			continue
		}
		for _, nested := range TokenizeReferences(field, variable.Value) {
			pending = append(pending, inspectReference(nested, context, dynamicNames))
		}
	}
	return inspection
}
