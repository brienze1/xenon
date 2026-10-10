package entity

type LocalVariables struct {
	Current             map[string]map[string]string
	ActiveEnvironmentID string
}

func NewLocalVariables() LocalVariables {
	return LocalVariables{Current: map[string]map[string]string{}}
}

func (l LocalVariables) HasActiveEnvironment() bool {
	return l.ActiveEnvironmentID != ""
}

func (l LocalVariables) WithActiveEnvironment(id string) LocalVariables {
	l.ActiveEnvironmentID = id
	return l
}

func (l LocalVariables) CurrentValues(scope ScopeRef) map[string]string {
	return l.Current[scope.String()]
}

func (l LocalVariables) LogValue() any {
	return map[string]any{"activeEnvironmentId": l.ActiveEnvironmentID, "scopes": len(l.Current)}
}
