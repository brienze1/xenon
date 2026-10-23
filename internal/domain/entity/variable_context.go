package entity

import (
	"slices"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/referencestatus"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
)

const (
	maxResolutionPasses  = 19
	vaultReferencePrefix = "vault:"
)

type VariableContext struct {
	Vault  map[string]string
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
	updated := NewVariableContext(append([]VariableLayer{{Scope: variablescope.Local, Variables: variables}}, v.Layers...)...)
	updated.Vault = v.Vault
	return updated
}

// WithIterationData puts the run's data row right after the send's script variables, before every stored scope.
func (v VariableContext) WithIterationData(row map[string]any) VariableContext {
	variables := make(map[string]ContextVariable, len(row))
	for name, value := range row {
		variables[name] = ContextVariable{Name: name, Value: LocalVariableText(value), Scope: variablescope.Data}
	}
	layers := make([]VariableLayer, 0, len(v.Layers)+1)
	inserted := false
	for _, layer := range v.Layers {
		if !inserted && layer.Scope != variablescope.Local {
			layers = append(layers, VariableLayer{Scope: variablescope.Data, Variables: variables})
			inserted = true
		}
		if layer.Scope != variablescope.Data {
			layers = append(layers, layer)
		}
	}
	if !inserted {
		layers = append(layers, VariableLayer{Scope: variablescope.Data, Variables: variables})
	}
	updated := NewVariableContext(layers...)
	updated.Vault = v.Vault
	return updated
}

// WithScriptScopes puts the scopes as the send's scripts left them over the stored values; unavailable secrets keep their state.
func (v VariableContext) WithScriptScopes(scopes ScriptScopes) VariableContext {
	layers := make([]VariableLayer, 0, len(v.Layers))
	for _, layer := range v.Layers {
		switch layer.Scope {
		case variablescope.Local:
			continue
		case variablescope.Data:
			layers = append(layers, layer)
		default:
			layers = append(layers, layer.withScriptScope(scopes.Scope(layer.Scope)))
		}
	}
	updated := NewVariableContext(layers...).WithLocal(scopes.Local.ValueMap())
	updated.Vault = v.Vault
	return updated
}

// WithVault makes {{vault:name}} resolve from these values (the workspace secrets as the send's scripts left them).
func (v VariableContext) WithVault(vault map[string]string) VariableContext {
	v.Vault = vault
	return v
}

func (v VariableContext) Lookup(name string) (ContextVariable, bool) {
	if secret, isVault := strings.CutPrefix(name, vaultReferencePrefix); isVault {
		return v.lookupVault(name, secret)
	}
	for _, layer := range v.Layers {
		if variable, found := layer.Lookup(name); found {
			return variable, true
		}
	}
	return ContextVariable{}, false
}

func (v VariableContext) lookupVault(reference, name string) (ContextVariable, bool) {
	if v.Vault != nil {
		value, found := v.Vault[name]
		return ContextVariable{Name: reference, Value: value, Scope: variablescope.Globals, Secret: true}, found
	}
	for _, layer := range v.Layers {
		if variable, found := layer.Lookup(name); found && layer.Scope == variablescope.Globals && variable.Secret {
			variable.Name = reference
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
	for _, value := range v.Vault {
		if value != "" && !slices.Contains(values, value) {
			values = append(values, value)
		}
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
