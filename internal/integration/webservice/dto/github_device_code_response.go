package dto

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type GitHubDeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (g *GitHubDeviceCodeResponse) ToEntity(sessionId string) entity.GitHubDeviceCode {
	return entity.GitHubDeviceCode{SessionID: sessionId, UserCode: g.UserCode, VerificationURI: g.VerificationURI, ExpiresIn: g.ExpiresIn, Interval: g.Interval}
}

func (g *GitHubDeviceCodeResponse) Session(sessionId string, now time.Time) entity.GitHubDeviceSession {
	return entity.NewGitHubDeviceSession(sessionId, g.DeviceCode, g.ExpiresIn, g.Interval, now)
}
