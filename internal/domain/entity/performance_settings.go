package entity

import "time"

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
		return invalidRunSetting("virtual_users", "must be between 1 and 100")
	case p.DurationSeconds < 1 || p.DurationSeconds > MaxDurationSeconds:
		return invalidRunSetting("duration_seconds", "must be between 1 and 3600")
	default:
		return p.Profile.Validate(p.VirtualUsers, p.DurationSeconds)
	}
}

func (p PerformanceSettings) WithProfileDefaults() PerformanceSettings {
	p.Profile = p.Profile.WithDefaults(p.VirtualUsers, p.DurationSeconds)
	return p
}

func (p PerformanceSettings) Duration() time.Duration {
	return time.Duration(p.DurationSeconds) * time.Second
}

func (p PerformanceSettings) ActiveUsersAt(elapsed time.Duration) int {
	return p.Profile.ActiveUsersAt(elapsed, p.VirtualUsers, p.DurationSeconds)
}
