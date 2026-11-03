package entity

import "maps"

type AISessionContext struct {
	Environments map[string]string
	SessionID    string
	LocalID      string
}

func NewAISessionContext(sessionId, localId, storedActiveEnvironmentId string) AISessionContext {
	return AISessionContext{SessionID: sessionId, Environments: map[string]string{}}.Enter(localId, storedActiveEnvironmentId)
}

// Enter makes localId the session's workspace, keeping the environment chosen there on an earlier visit.
func (a AISessionContext) Enter(localId, storedActiveEnvironmentId string) AISessionContext {
	a.LocalID = localId
	if !a.HasWorkspace() || a.HasEntered(localId) {
		return a
	}
	return a.withEnvironment(storedActiveEnvironmentId)
}

func (a AISessionContext) HasEntered(localId string) bool {
	_, entered := a.Environments[localId]
	return entered
}

func (a AISessionContext) HasWorkspace() bool {
	return a.LocalID != ""
}

func (a AISessionContext) ChooseEnvironment(environmentId string) AISessionContext {
	if !a.HasWorkspace() {
		return a
	}
	return a.withEnvironment(environmentId)
}

// EnvironmentID is the environment chosen for the session's workspace, or NoEnvironmentID; never empty.
func (a AISessionContext) EnvironmentID() string {
	if environmentId, entered := a.Environments[a.LocalID]; entered && a.HasWorkspace() {
		return environmentId
	}
	return NoEnvironmentID
}

func (a AISessionContext) HasEnvironment() bool {
	return !UsesNoEnvironment(a.EnvironmentID())
}

// ActiveEnvironmentID is the environment chosen for the session's workspace, empty when none.
func (a AISessionContext) ActiveEnvironmentID() string {
	if a.HasEnvironment() {
		return a.EnvironmentID()
	}
	return ""
}

// InRegistry leaves a workspace that is no longer registered.
func (a AISessionContext) InRegistry(registry *AppRegistry) AISessionContext {
	if _, found := registry.Find(a.LocalID); a.HasWorkspace() && !found {
		return a.Leave()
	}
	return a
}

func (a AISessionContext) Leave() AISessionContext {
	a.LocalID = ""
	return a
}

func (a AISessionContext) LogValue() any {
	return map[string]any{"sessionId": a.SessionID, "localId": a.LocalID, "environmentId": a.EnvironmentID()}
}

func (a AISessionContext) withEnvironment(environmentId string) AISessionContext {
	environments := make(map[string]string, len(a.Environments)+1)
	maps.Copy(environments, a.Environments)
	environments[a.LocalID] = sessionEnvironmentID(environmentId)
	a.Environments = environments
	return a
}

func sessionEnvironmentID(environmentId string) string {
	if UsesActiveEnvironment(environmentId) {
		return NoEnvironmentID
	}
	return environmentId
}
