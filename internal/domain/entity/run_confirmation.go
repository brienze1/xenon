package entity

type RunConfirmation struct {
	Target      RunTarget
	Environment Environment
}

func (r RunConfirmation) Approval(client string, config RunConfig) AIApproval {
	return AIApproval{
		Client:      client,
		Summary:     NewRunApprovalSummary(r.Target.Name, len(config.SelectedRows())),
		Environment: NewAIApprovalEnvironment(r.Environment.ID, r.Environment.Name, r.Environment.Color.String()),
	}
}

func (r RunConfirmation) Confirmed(start RunStart) RunStart {
	start.ConfirmedEnvironmentID = r.Environment.ID
	return start
}
