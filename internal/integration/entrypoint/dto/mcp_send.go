package dto

import "github.com/brienze1/xenon/internal/domain/enums/historysource"

type MCPSend struct {
	MCPSendTarget
	SendUnresolved   bool `json:"send_unresolved"`
	SendExpiredToken bool `json:"send_expired_token"`
}

func (m *MCPSend) SendRequest(client, localId, tabId, confirmedEnvironmentId string, request Request) SendRequest {
	return NewSendRequest(NewSendOrigin(historysource.AI, client), localId, tabId, confirmedEnvironmentId, request, m.SendUnresolved, m.SendExpiredToken)
}
