package entity

type AIAccessPatch struct {
	Enabled       *bool
	Port          *int
	SkipApprovals *bool
}

func (p AIAccessPatch) ApplyTo(settings AIAccessSettings) AIAccessSettings {
	if p.Enabled != nil {
		settings.Enabled = *p.Enabled
	}
	if p.Port != nil {
		settings.Port = *p.Port
	}
	if p.SkipApprovals != nil {
		settings.SkipApprovals = *p.SkipApprovals
	}
	return settings
}
