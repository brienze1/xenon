package entity

import "slices"

type SendPreflight struct {
	ConfirmEnvironment *Environment
	MissingSecrets     []string
	LockedSecrets      []string
}

func (s SendPreflight) NeedsConfirmation() bool {
	return s.ConfirmEnvironment != nil
}

func (s SendPreflight) HasUnavailableSecrets() bool {
	return s.HasMissingSecrets() || s.HasLockedSecrets()
}

func (s SendPreflight) HasMissingSecrets() bool {
	return len(s.MissingSecrets) > 0
}

func (s SendPreflight) HasLockedSecrets() bool {
	return len(s.LockedSecrets) > 0
}

// WithConfirmation asks for a confirmation when the active environment requires one; nil is "No environment".
func (s SendPreflight) WithConfirmation(environment *Environment) SendPreflight {
	if environment != nil && environment.ConfirmBeforeSend {
		s.ConfirmEnvironment = environment
	}
	return s
}

func (s SendPreflight) AwaitsConfirmation(acknowledgement SendAcknowledgement) bool {
	return s.NeedsConfirmation() && !acknowledgement.Confirms(*s.ConfirmEnvironment)
}

func (s SendPreflight) AwaitsSecretsAcknowledgement(acknowledgement SendAcknowledgement) bool {
	return s.HasUnavailableSecrets() && !acknowledgement.MissingSecrets
}

func (s SendPreflight) IsClear() bool {
	return !s.NeedsConfirmation() && !s.HasUnavailableSecrets()
}

func (s SendPreflight) WithUnavailable(variable ContextVariable) SendPreflight {
	if variable.Locked {
		s.LockedSecrets = withName(s.LockedSecrets, variable.Name)
	} else {
		s.MissingSecrets = withName(s.MissingSecrets, variable.Name)
	}
	return s
}

func withName(names []string, name string) []string {
	if slices.Contains(names, name) {
		return names
	}
	sorted := append(append([]string{}, names...), name)
	slices.Sort(sorted)
	return sorted
}
