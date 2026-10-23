package entity

type RunStart struct {
	LocalID                   string
	ItemID                    string
	ConfirmedEnvironmentID    string
	Config                    RunConfig
	AcknowledgeMissingSecrets bool
}

func NewRunStart(localId, itemId string, config RunConfig) RunStart {
	return RunStart{LocalID: localId, ItemID: itemId, Config: config}
}

// Acknowledgement always accepts expired tokens, as Postman's runner does.
func (r RunStart) Acknowledgement() SendAcknowledgement {
	return SendAcknowledgement{ConfirmedEnvironmentID: r.ConfirmedEnvironmentID, MissingSecrets: r.AcknowledgeMissingSecrets, ExpiredToken: true}
}
