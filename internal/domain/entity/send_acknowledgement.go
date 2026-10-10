package entity

type SendAcknowledgement struct {
	ConfirmedEnvironmentID string
	MissingSecrets         bool
}

func (s SendAcknowledgement) Confirms(environment Environment) bool {
	return s.ConfirmedEnvironmentID == environment.ID
}

func (s SendAcknowledgement) LogValue() any {
	return map[string]any{"confirmedEnvironmentId": s.ConfirmedEnvironmentID, "missingSecrets": s.MissingSecrets}
}
