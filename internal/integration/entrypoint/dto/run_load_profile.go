package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/loadprofilekind"
)

type RunLoadProfile struct {
	Kind         string `json:"kind"`
	InitialUsers int    `json:"initial_users"`
	RampSeconds  int    `json:"ramp_seconds"`
	BaseUsers    int    `json:"base_users"`
	SpikeSeconds int    `json:"spike_seconds"`
	PeakSeconds  int    `json:"peak_seconds"`
}

func NewRunLoadProfile(profile entity.LoadProfile) RunLoadProfile {
	return RunLoadProfile{
		Kind:         profile.Kind.String(),
		InitialUsers: profile.InitialUsers,
		RampSeconds:  profile.RampSeconds,
		BaseUsers:    profile.BaseUsers,
		SpikeSeconds: profile.SpikeSeconds,
		PeakSeconds:  profile.PeakSeconds,
	}
}

func (r *RunLoadProfile) ToEntity() entity.LoadProfile {
	return entity.LoadProfile{
		Kind:         loadProfileKind(r.Kind),
		InitialUsers: r.InitialUsers,
		RampSeconds:  r.RampSeconds,
		BaseUsers:    r.BaseUsers,
		SpikeSeconds: r.SpikeSeconds,
		PeakSeconds:  r.PeakSeconds,
	}
}

func loadProfileKind(kind string) enums.LoadProfileKind {
	if kind == "" {
		return loadprofilekind.Fixed
	}
	return loadprofilekind.GetLoadProfileKind(kind)
}
