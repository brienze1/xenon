package entity

import (
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/referencestatus"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

const maxResolutionPasses = 19

type VariableContext struct {
	Layers []VariableLayer
}

func NewVariableContext(layers ...VariableLayer) VariableContext {
	return VariableContext{Layers: append([]VariableLayer{}, layers...)}
}

// WithLocal puts the send's script variables (pm.variables) in front of every other scope.
func (v VariableContext) WithLocal(values map[string]any) VariableContext {
	variables := make(map[string]ContextVariable, len(values))
	for name, value := range values {
		variables[name] = ContextVariable{Name: name, Value: LocalVariableText(value), Scope: variablescope.Local}
	}
	return NewVariableContext(append([]VariableLayer{{Scope: variablescope.Local, Variables: variables}}, v.Layers...)...)
}

func (v VariableContext) Lookup(name string) (ContextVariable, bool) {
	for _, layer := range v.Layers {
		if variable, found := layer.Lookup(name); found {
			return variable, true
		}
	}
	return ContextVariable{}, false
}

func (v VariableContext) SecretValues() []string {
	values := make([]string, 0)
	for _, layer := range v.Layers {
		values = append(values, layer.secretValues()...)
	}
	return values
}

func (v VariableContext) Masked(text string) string {
	for _, secret := range v.SecretValues() {
		text = strings.ReplaceAll(text, secret, MaskedValue)
	}
	return text
}

// Variables lists every variable once, strongest scope first (a name defined in several scopes keeps the strongest).
func (v VariableContext) Variables() []ContextVariable {
	seen := make(map[string]bool)
	variables := make([]ContextVariable, 0)
	for _, layer := range v.Layers {
		for _, variable := range layer.sortedVariables() {
			if !seen[variable.Name] {
				seen[variable.Name] = true
				variables = append(variables, variable)
			}
		}
	}
	return variables
}

// Resolve replaces references in up to 19 passes (Postman's limit); unknown names stay literal.
func (v VariableContext) Resolve(text string, dynamic DynamicValue) string {
	for passesLeft := maxResolutionPasses; passesLeft > 0; passesLeft-- {
		next := v.resolvePass(text, dynamic)
		if next == text {
			return text
		}
		text = next
	}
	return text
}

// UnavailableSecretsIn lists the secrets without a value or in a locked vault that the text references, nested values followed.
func (v VariableContext) UnavailableSecretsIn(text string) []ContextVariable {
	unavailable := make([]ContextVariable, 0)
	for passesLeft := maxResolutionPasses; passesLeft > 0; passesLeft-- {
		for _, reference := range TokenizeReferences("", text) {
			if variable, found := v.Lookup(reference.Name); found && variable.IsUnavailableSecret() {
				unavailable = append(unavailable, variable)
			}
		}
		next := v.resolvePass(text, nil)
		if next == text {
			return unavailable
		}
		text = next
	}
	return unavailable
}

// Inspect resolves the reference without generating dynamic values; secrets are masked.
func (v VariableContext) Inspect(reference Reference, dynamicNames []string) ResolvedReference {
	variable, found := v.Lookup(reference.Name)
	switch {
	case found && variable.IsMissingSecret():
		return ResolvedReference{Reference: reference, Status: referencestatus.MissingSecret, Scope: variable.Scope, Secret: true}
	case found && variable.IsLockedSecret():
		return ResolvedReference{Reference: reference, Status: referencestatus.LockedSecret, Scope: variable.Scope, Secret: true}
	case found:
		return v.resolvedReference(reference, variable)
	case slices.Contains(dynamicNames, reference.Name):
		return ResolvedReference{Reference: reference, Status: referencestatus.Dynamic, Scope: variablescope.Dynamic}
	default:
		return ResolvedReference{Reference: reference, Status: referencestatus.Unresolved, Scope: variablescope.None}
	}
}

func (v VariableContext) resolvedReference(reference Reference, variable ContextVariable) ResolvedReference {
	resolved := v.Resolve(reference.Placeholder(), nil)
	if v.resolvePass(resolved, nil) != resolved {
		return ResolvedReference{Reference: reference, Status: referencestatus.Unresolved, Scope: variable.Scope, Secret: variable.Secret}
	}
	return ResolvedReference{Reference: reference, Status: referencestatus.Resolved, Scope: variable.Scope, Secret: variable.Secret, Value: v.Masked(resolved)}
}

func (v VariableContext) resolvePass(text string, dynamic DynamicValue) string {
	return variableReference.ReplaceAllStringFunc(text, func(match string) string {
		if value, found := v.lookupValue(match[2:len(match)-2], dynamic); found {
			return value
		}
		return match
	})
}

func (v VariableContext) lookupValue(name string, dynamic DynamicValue) (string, bool) {
	if variable, found := v.Lookup(name); found {
		return variable.Value, variable.IsResolvable()
	}
	if dynamic == nil {
		return "", false
	}
	return dynamic(name)
}
