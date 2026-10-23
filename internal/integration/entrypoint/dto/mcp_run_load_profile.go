package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPRunLoadProfile struct {
	Kind         *string `json:"kind" description:"Profile kind." enum:"fixed,ramp_up,spike,peak"`
	InitialUsers *int    `json:"initial_users" description:"Ramp up start users."`
	RampSeconds  *int    `json:"ramp_seconds" description:"Ramp up seconds."`
	BaseUsers    *int    `json:"base_users" description:"Spike or peak base users."`
	SpikeSeconds *int    `json:"spike_seconds" description:"Spike seconds."`
	PeakSeconds  *int    `json:"peak_seconds" description:"Peak seconds."`
}

func (m *MCPRunLoadProfile) ToEntity() entity.LoadProfileOverrides {
	if m == nil {
		return entity.LoadProfileOverrides{}
	}
	overrides := entity.LoadProfileOverrides{
		InitialUsers: m.InitialUsers,
		RampSeconds:  m.RampSeconds,
		BaseUsers:    m.BaseUsers,
		SpikeSeconds: m.SpikeSeconds,
		PeakSeconds:  m.PeakSeconds,
	}
	if m.Kind != nil {
		kind := loadProfileKind(*m.Kind)
		overrides.Kind = &kind
	}
	return overrides
}
