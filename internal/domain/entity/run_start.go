package entity

type RunStart struct {
	LocalID                   string
	ItemID                    string
	ConfirmedEnvironmentID    string
	Config                    RunConfig
	AcknowledgeMissingSecrets bool
}

// Acknowledgement is what every row of the run sends with: expired tokens are always accepted, as in Postman's runner.
func (r RunStart) Acknowledgement() SendAcknowledgement {
	return SendAcknowledgement{ConfirmedEnvironmentID: r.ConfirmedEnvironmentID, MissingSecrets: r.AcknowledgeMissingSecrets, ExpiredToken: true}
}
