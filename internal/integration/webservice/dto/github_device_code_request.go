package dto

import "net/url"

const gitHubScopes = "repo read:user user:email write:public_key"

type GitHubDeviceCodeRequest struct {
	ClientID string
	Scope    string
}

func NewGitHubDeviceCodeRequest(clientId string) GitHubDeviceCodeRequest {
	return GitHubDeviceCodeRequest{ClientID: clientId, Scope: gitHubScopes}
}

func (g GitHubDeviceCodeRequest) Form() string {
	return url.Values{"client_id": {g.ClientID}, "scope": {g.Scope}}.Encode()
}
