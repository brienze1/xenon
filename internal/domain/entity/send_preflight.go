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
	return len(s.MissingSecrets)+len(s.LockedSecrets) > 0
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
