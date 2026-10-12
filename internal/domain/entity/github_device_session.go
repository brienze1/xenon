package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/signinstatus"
)

const gitHubSlowDownStep = 5 * time.Second

type GitHubDeviceSession struct {
	ExpiresAt  time.Time
	NextPollAt time.Time
	ID         string
	DeviceCode string
	Interval   time.Duration
}

func NewGitHubDeviceSession(id, deviceCode string, expiresIn, interval int, now time.Time) GitHubDeviceSession {
	return GitHubDeviceSession{
		ID:         id,
		DeviceCode: deviceCode,
		ExpiresAt:  now.Add(time.Duration(expiresIn) * time.Second),
		Interval:   time.Duration(interval) * time.Second,
		NextPollAt: now.Add(time.Duration(interval) * time.Second),
	}
}

func (g GitHubDeviceSession) IsExpired(now time.Time) bool {
	return !now.Before(g.ExpiresAt)
}

func (g GitHubDeviceSession) IsTooEarly(now time.Time) bool {
	return now.Before(g.NextPollAt)
}

func (g GitHubDeviceSession) Polled(now time.Time) GitHubDeviceSession {
	g.NextPollAt = now.Add(g.Interval)
	return g
}

func (g GitHubDeviceSession) SlowedDown(now time.Time) GitHubDeviceSession {
	g.Interval += gitHubSlowDownStep
	return g.Polled(now)
}

func (g GitHubDeviceSession) IntervalSeconds() int {
	return int(g.Interval / time.Second)
}

func (g GitHubDeviceSession) LogValue() any {
	return map[string]any{"id": g.ID, "interval": g.Interval.String(), "expiresAt": g.ExpiresAt, "nextPollAt": g.NextPollAt}
}

func (g GitHubDeviceSession) ExpiredSignIn() GitHubSignIn {
	return GitHubSignIn{Status: signinstatus.Expired, Interval: g.IntervalSeconds()}
}

func (g GitHubDeviceSession) PendingSignIn() GitHubSignIn {
	return GitHubSignIn{Status: signinstatus.Pending, Interval: g.IntervalSeconds()}
}
