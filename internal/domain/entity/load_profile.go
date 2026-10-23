package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/loadprofilekind"
)

type LoadProfile struct {
	Kind         enums.LoadProfileKind
	InitialUsers int
	RampSeconds  int
	BaseUsers    int
	SpikeSeconds int
	PeakSeconds  int
}

func NewFixedLoadProfile() LoadProfile {
	return LoadProfile{Kind: loadprofilekind.Fixed}
}

func (l LoadProfile) Validate(virtualUsers, durationSeconds int) error {
	switch l.Kind {
	case loadprofilekind.Fixed:
		return nil
	case loadprofilekind.RampUp:
		return l.validateRampUp(virtualUsers, durationSeconds)
	case loadprofilekind.Spike:
		return l.validateSpike(virtualUsers, durationSeconds)
	case loadprofilekind.Peak:
		return l.validatePeak(virtualUsers, durationSeconds)
	default:
		return invalidRunSetting("profile.kind", "must be fixed, ramp_up, spike or peak")
	}
}

func (l LoadProfile) validateRampUp(virtualUsers, durationSeconds int) error {
	switch {
	case l.InitialUsers < 1 || l.InitialUsers > virtualUsers:
		return invalidRunSetting("profile.initial_users", "must be between 1 and the virtual users")
	case l.RampSeconds < 1 || l.RampSeconds > durationSeconds:
		return invalidRunSetting("profile.ramp_seconds", "must be between 1 and the duration")
	default:
		return nil
	}
}

func (l LoadProfile) validateSpike(virtualUsers, durationSeconds int) error {
	switch {
	case l.BaseUsers < 1 || l.BaseUsers > virtualUsers:
		return invalidRunSetting("profile.base_users", "must be between 1 and the virtual users")
	case l.SpikeSeconds < 1 || l.SpikeSeconds > durationSeconds:
		return invalidRunSetting("profile.spike_seconds", "must be between 1 and the duration")
	default:
		return nil
	}
}

func (l LoadProfile) validatePeak(virtualUsers, durationSeconds int) error {
	switch {
	case l.BaseUsers < 1 || l.BaseUsers > virtualUsers:
		return invalidRunSetting("profile.base_users", "must be between 1 and the virtual users")
	case l.PeakSeconds < 1 || l.PeakSeconds >= durationSeconds:
		return invalidRunSetting("profile.peak_seconds", "must be at least 1 and shorter than the duration")
	default:
		return nil
	}
}
