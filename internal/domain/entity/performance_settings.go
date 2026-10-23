package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
)

const (
	DefaultVirtualUsers    = 10
	MaxVirtualUsers        = 100
	DefaultDurationSeconds = 600
	MaxDurationSeconds     = 3600
)

type PerformanceSettings struct {
	DataFile        string
	Profile         LoadProfile
	VirtualUsers    int
	DurationSeconds int
}

func NewPerformanceSettings() PerformanceSettings {
	return PerformanceSettings{VirtualUsers: DefaultVirtualUsers, DurationSeconds: DefaultDurationSeconds, Profile: NewFixedLoadProfile()}
}

func (p PerformanceSettings) Validate() error {
	switch {
	case p.VirtualUsers < 1 || p.VirtualUsers > MaxVirtualUsers:
		return invalidRunSetting("virtual_users", "must be between 1 and 100", "RUN-04400")
	case p.DurationSeconds < 1 || p.DurationSeconds > MaxDurationSeconds:
		return invalidRunSetting("duration_seconds", "must be between 1 and 3600", "RUN-04400")
	default:
		return p.Profile.Validate(p.VirtualUsers, p.DurationSeconds)
	}
}

func (p PerformanceSettings) Override(overrides PerformanceOverrides) PerformanceSettings {
	p.VirtualUsers = overridden(overrides.VirtualUsers, p.VirtualUsers)
	p.DurationSeconds = overridden(overrides.DurationSeconds, p.DurationSeconds)
	p.DataFile = overridden(overrides.DataFile, p.DataFile)
	p.Profile = p.Profile.Override(overrides.Profile)
	return p
}

func (p PerformanceSettings) WithProfileDefaults() PerformanceSettings {
	p.Profile = p.Profile.WithDefaults(p.VirtualUsers, p.DurationSeconds)
	return p
}

func (p PerformanceSettings) Duration() time.Duration {
	return time.Duration(p.DurationSeconds) * time.Second
}

func (p PerformanceSettings) HasElapsed(elapsed time.Duration) bool {
	return elapsed >= p.Duration()
}

func (p PerformanceSettings) ActiveUsersAt(elapsed time.Duration) int {
	return p.Profile.ActiveUsersAt(elapsed, p.VirtualUsers, p.DurationSeconds)
}

// ReportedSeconds is the whole duration of a completed run, else the seconds it ran, rounded up and capped at the duration.
func (p PerformanceSettings) ReportedSeconds(status enums.RunStatus, elapsed time.Duration) int {
	if status == runstatus.Completed {
		return p.DurationSeconds
	}
	return min(p.DurationSeconds, int((elapsed+time.Second-1)/time.Second))
}

// SecondAt is the report second elapsed falls in, capped at the run's last second.
func (p PerformanceSettings) SecondAt(elapsed time.Duration) int {
	return min(int(elapsed/time.Second), p.DurationSeconds-1)
}
