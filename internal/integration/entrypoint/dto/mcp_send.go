package dto

import "github.com/brienze1/xenon/internal/domain/enums/historysource"

type MCPSend struct {
	MCPSendTarget
	SendUnresolved   bool `json:"send_unresolved" description:"Send even when a variable or secret is unresolved."`
	SendExpiredToken bool `json:"send_expired_token" description:"Send even when the OAuth token expired and could not be refreshed."`
}

func (m *MCPSend) SendRequest(client, localId, tabId, confirmedEnvironmentId string, request Request) SendRequest {
	return NewSendRequest(NewSendOrigin(historysource.AI, client), localId, tabId, confirmedEnvironmentId, request, m.SendUnresolved, m.SendExpiredToken)
}
