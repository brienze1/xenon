package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/loadprofilekind"
)

type LoadProfile struct {
	Kind         string `json:"kind"`
	InitialUsers int    `json:"initial_users"`
	RampSeconds  int    `json:"ramp_seconds"`
	BaseUsers    int    `json:"base_users"`
	SpikeSeconds int    `json:"spike_seconds"`
	PeakSeconds  int    `json:"peak_seconds"`
}

func NewLoadProfile(profile entity.LoadProfile) LoadProfile {
	return LoadProfile{
		Kind:         profile.Kind.String(),
		InitialUsers: profile.InitialUsers,
		RampSeconds:  profile.RampSeconds,
		BaseUsers:    profile.BaseUsers,
		SpikeSeconds: profile.SpikeSeconds,
		PeakSeconds:  profile.PeakSeconds,
	}
}

func (l *LoadProfile) ToEntity() entity.LoadProfile {
	return entity.LoadProfile{
		Kind:         loadprofilekind.GetLoadProfileKind(l.Kind),
		InitialUsers: l.InitialUsers,
		RampSeconds:  l.RampSeconds,
		BaseUsers:    l.BaseUsers,
		SpikeSeconds: l.SpikeSeconds,
		PeakSeconds:  l.PeakSeconds,
	}
}
