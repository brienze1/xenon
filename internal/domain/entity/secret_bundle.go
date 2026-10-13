package entity

type SecretBundle struct {
	Scopes map[string]map[string]string
}

func NewSecretBundle(scopes map[string]map[string]string) SecretBundle {
	bundle := SecretBundle{Scopes: map[string]map[string]string{}}
	for scope, values := range scopes {
		bundle = bundle.With(ParseScopeRef(scope), values)
	}
	return bundle
}

func (s SecretBundle) With(scope ScopeRef, values map[string]string) SecretBundle {
	scopes := copyScopes(s.Scopes)
	if len(values) > 0 {
		scopes[scope.String()] = copyValues(values)
	}
	s.Scopes = scopes
	return s
}

func (s SecretBundle) IsEmpty() bool {
	return len(s.Scopes) == 0
}

func (s SecretBundle) For(scope ScopeRef) map[string]string {
	return copyValues(s.Scopes[scope.String()])
}

func (s SecretBundle) LogValue() any {
	counts := make(map[string]int, len(s.Scopes))
	for scope, values := range s.Scopes {
		counts[scope] = len(values)
	}
	return map[string]any{"scopes": counts}
}
