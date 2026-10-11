package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type GitHubDeviceCode struct {
	SessionID       string `json:"session_id"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func NewGitHubDeviceCode(code entity.GitHubDeviceCode) GitHubDeviceCode {
	return GitHubDeviceCode{SessionID: code.SessionID, UserCode: code.UserCode, VerificationURI: code.VerificationURI, ExpiresIn: code.ExpiresIn, Interval: code.Interval}
}
