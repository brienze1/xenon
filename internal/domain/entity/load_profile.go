package entity

import (
	"math"
	"time"

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
		return invalidRunSetting("profile.kind", "must be "+loadprofilekind.Fixed.String()+", "+loadprofilekind.RampUp.String()+", "+loadprofilekind.Spike.String()+" or "+loadprofilekind.Peak.String(), "RUN-04400")
	}
}

func (l LoadProfile) validateRampUp(virtualUsers, durationSeconds int) error {
	switch {
	case l.InitialUsers < 1 || l.InitialUsers > virtualUsers:
		return invalidRunSetting("profile.initial_users", "must be between 1 and the virtual users", "RUN-04400")
	case l.RampSeconds < 1 || l.RampSeconds > durationSeconds:
		return invalidRunSetting("profile.ramp_seconds", "must be between 1 and the duration", "RUN-04400")
	default:
		return nil
	}
}

func (l LoadProfile) validateSpike(virtualUsers, durationSeconds int) error {
	switch {
	case l.BaseUsers < 1 || l.BaseUsers > virtualUsers:
		return invalidRunSetting("profile.base_users", "must be between 1 and the virtual users", "RUN-04400")
	case l.SpikeSeconds < 1 || l.SpikeSeconds > durationSeconds:
		return invalidRunSetting("profile.spike_seconds", "must be between 1 and the duration", "RUN-04400")
	default:
		return nil
	}
}

func (l LoadProfile) validatePeak(virtualUsers, durationSeconds int) error {
	switch {
	case l.BaseUsers < 1 || l.BaseUsers > virtualUsers:
		return invalidRunSetting("profile.base_users", "must be between 1 and the virtual users", "RUN-04400")
	case l.PeakSeconds < 1 || l.PeakSeconds >= durationSeconds:
		return invalidRunSetting("profile.peak_seconds", "must be at least 1 and shorter than the duration", "RUN-04400")
	default:
		return nil
	}
}

func (l LoadProfile) Override(overrides LoadProfileOverrides) LoadProfile {
	l.Kind = overridden(overrides.Kind, l.Kind)
	l.InitialUsers = overridden(overrides.InitialUsers, l.InitialUsers)
	l.RampSeconds = overridden(overrides.RampSeconds, l.RampSeconds)
	l.BaseUsers = overridden(overrides.BaseUsers, l.BaseUsers)
	l.SpikeSeconds = overridden(overrides.SpikeSeconds, l.SpikeSeconds)
	l.PeakSeconds = overridden(overrides.PeakSeconds, l.PeakSeconds)
	return l
}

// WithDefaults fills the parameters of the kind left at zero with the FR-011 defaults.
func (l LoadProfile) WithDefaults(virtualUsers, durationSeconds int) LoadProfile {
	tenthOfUsers := max(1, virtualUsers/10)
	switch l.Kind {
	case loadprofilekind.RampUp:
		l.InitialUsers = defaulted(l.InitialUsers, 1)
		l.RampSeconds = defaulted(l.RampSeconds, max(1, durationSeconds/4))
	case loadprofilekind.Spike:
		l.BaseUsers = defaulted(l.BaseUsers, tenthOfUsers)
		l.SpikeSeconds = defaulted(l.SpikeSeconds, max(1, durationSeconds/10))
	case loadprofilekind.Peak:
		l.BaseUsers = defaulted(l.BaseUsers, tenthOfUsers)
		l.PeakSeconds = defaulted(l.PeakSeconds, max(1, durationSeconds/2))
	}
	return l
}

func (l LoadProfile) ActiveUsersAt(elapsed time.Duration, virtualUsers, durationSeconds int) int {
	at := elapsed.Seconds()
	switch l.Kind {
	case loadprofilekind.RampUp:
		return l.rampUpUsersAt(at, virtualUsers)
	case loadprofilekind.Spike:
		return l.spikeUsersAt(at, virtualUsers, durationSeconds)
	case loadprofilekind.Peak:
		return l.peakUsersAt(at, virtualUsers, durationSeconds)
	default:
		return virtualUsers
	}
}

func (l LoadProfile) rampUpUsersAt(at float64, virtualUsers int) int {
	ramp := float64(l.RampSeconds)
	if at >= ramp {
		return virtualUsers
	}
	return interpolatedUsers(l.InitialUsers, virtualUsers, at/ramp)
}

func (l LoadProfile) spikeUsersAt(at float64, virtualUsers, durationSeconds int) int {
	start := float64(durationSeconds-l.SpikeSeconds) / 2
	if at >= start && at < start+float64(l.SpikeSeconds) {
		return virtualUsers
	}
	return l.BaseUsers
}

func (l LoadProfile) peakUsersAt(at float64, virtualUsers, durationSeconds int) int {
	ramp := float64(durationSeconds-l.PeakSeconds) / 2
	holdEnd := ramp + float64(l.PeakSeconds)
	switch {
	case at < ramp:
		return interpolatedUsers(l.BaseUsers, virtualUsers, at/ramp)
	case at < holdEnd:
		return virtualUsers
	case at >= holdEnd+ramp:
		return l.BaseUsers
	default:
		return interpolatedUsers(virtualUsers, l.BaseUsers, (at-holdEnd)/ramp)
	}
}

func interpolatedUsers(from, to int, fraction float64) int {
	return from + int(math.Round(float64(to-from)*fraction))
}

func defaulted(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}
