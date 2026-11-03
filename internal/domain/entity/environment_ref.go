package entity

// ResolveEnvironmentRef resolves an environment name or id, the active environment or NoEnvironmentID; false when no environment matches.
func ResolveEnvironmentRef(environments []Environment, ref string) (string, bool) {
	if UsesActiveEnvironment(ref) || UsesNoEnvironment(ref) {
		return ref, true
	}
	for _, environment := range environments {
		if environment.ID == ref || environment.Name == ref {
			return environment.ID, true
		}
	}
	return "", false
}
