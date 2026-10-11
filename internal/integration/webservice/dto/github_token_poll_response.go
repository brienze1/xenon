package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/signinstatus"
)

const (
	gitHubSlowDown     = "slow_down"
	gitHubPending      = "authorization_pending"
	gitHubExpiredToken = "expired_token"
	gitHubAccessDenied = "access_denied"
)

type GitHubTokenPollResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
	Interval    int    `json:"interval"`
}

func (g *GitHubTokenPollResponse) IsSlowDown() bool {
	return g.Error == gitHubSlowDown
}

func (g *GitHubTokenPollResponse) Status() enums.SignInStatus {
	switch {
	case g.AccessToken != "":
		return signinstatus.Done
	case g.Error == gitHubPending, g.IsSlowDown():
		return signinstatus.Pending
	case g.Error == gitHubExpiredToken:
		return signinstatus.Expired
	case g.Error == gitHubAccessDenied:
		return signinstatus.Denied
	default:
		return signinstatus.Unknown
	}
}

func (g *GitHubTokenPollResponse) ToEntity(session entity.GitHubDeviceSession) entity.GitHubSignIn {
	return entity.GitHubSignIn{Status: g.Status(), Token: g.AccessToken, Interval: session.IntervalSeconds()}
}
