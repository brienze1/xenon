package dto

import "net/url"

const deviceCodeGrantType = "urn:ietf:params:oauth:grant-type:device_code"

type GitHubTokenPollRequest struct {
	ClientID   string
	DeviceCode string
}

func NewGitHubTokenPollRequest(clientId, deviceCode string) GitHubTokenPollRequest {
	return GitHubTokenPollRequest{ClientID: clientId, DeviceCode: deviceCode}
}

func (g GitHubTokenPollRequest) Form() string {
	return url.Values{"client_id": {g.ClientID}, "device_code": {g.DeviceCode}, "grant_type": {deviceCodeGrantType}}.Encode()
}
