package entity

func copyValues(values map[string]string) map[string]string {
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func copyScopes(scopes map[string]map[string]string) map[string]map[string]string {
	copied := make(map[string]map[string]string, len(scopes))
	for scope, values := range scopes {
		copied[scope] = copyValues(values)
	}
	return copied
}

func withScopeValues(scopes map[string]map[string]string, scope ScopeRef, values map[string]string) map[string]map[string]string {
	copied := copyScopes(scopes)
	delete(copied, scope.String())
	if len(values) > 0 {
		copied[scope.String()] = copyValues(values)
	}
	return copied
}
