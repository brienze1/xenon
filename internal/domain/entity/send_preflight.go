package entity

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
