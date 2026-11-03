package dto

import "github.com/brienze1/xenon/internal/domain/enums/historysource"

type MCPSend struct {
	MCPSendTarget
	Environment      string `json:"environment" description:"Environment name or id, or none, for this send only."`
	SendUnresolved   bool   `json:"send_unresolved" description:"Send even when a variable or secret is unresolved."`
	SendExpiredToken bool   `json:"send_expired_token" description:"Send even when the OAuth token expired and could not be refreshed."`
}

func (m *MCPSend) OverridesEnvironment() bool {
	return m.Environment != ""
}

func (m *MCPSend) SendRequest(client, session, localId, tabId, environmentId, confirmedEnvironmentId string, request Request) SendRequest {
	return NewSendRequest(NewSendOrigin(historysource.AI, client, session), localId, tabId, confirmedEnvironmentId, request, m.SendUnresolved, m.SendExpiredToken).WithAlias(m.AliasId).InEnvironment(environmentId)
}
