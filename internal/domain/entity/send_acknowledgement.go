package entity

type SendAcknowledgement struct {
	ConfirmedEnvironmentID string
	MissingSecrets         bool
	ExpiredToken           bool
}

func (s SendAcknowledgement) Confirms(environment Environment) bool {
	return s.ConfirmedEnvironmentID == environment.ID
}

func (s SendAcknowledgement) AcceptsExpiredToken() bool {
	return s.ExpiredToken
}

func (s SendAcknowledgement) LogValue() any {
	return map[string]any{"confirmedEnvironmentId": s.ConfirmedEnvironmentID, "missingSecrets": s.MissingSecrets, "expiredToken": s.ExpiredToken}
}
