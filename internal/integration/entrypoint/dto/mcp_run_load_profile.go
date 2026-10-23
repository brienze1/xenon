package dto

type MCPRunLoadProfile struct {
	Kind         *string `json:"kind" description:"Profile kind." enum:"fixed,ramp_up,spike,peak"`
	InitialUsers *int    `json:"initial_users" description:"Ramp up start users."`
	RampSeconds  *int    `json:"ramp_seconds" description:"Ramp up seconds."`
	BaseUsers    *int    `json:"base_users" description:"Spike or peak base users."`
	SpikeSeconds *int    `json:"spike_seconds" description:"Spike seconds."`
	PeakSeconds  *int    `json:"peak_seconds" description:"Peak seconds."`
}

func (m *MCPRunLoadProfile) Over(profile RunLoadProfile) RunLoadProfile {
	if m == nil {
		return profile
	}
	profile.Kind = valueOr(m.Kind, profile.Kind)
	profile.InitialUsers = valueOr(m.InitialUsers, profile.InitialUsers)
	profile.RampSeconds = valueOr(m.RampSeconds, profile.RampSeconds)
	profile.BaseUsers = valueOr(m.BaseUsers, profile.BaseUsers)
	profile.SpikeSeconds = valueOr(m.SpikeSeconds, profile.SpikeSeconds)
	profile.PeakSeconds = valueOr(m.PeakSeconds, profile.PeakSeconds)
	return profile
}
