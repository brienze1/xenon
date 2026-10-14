package dto

type MCPSession struct {
	Id string `json:"id"`
}

func NewMCPSession(id string) MCPSession {
	return MCPSession{Id: id}
}

func (m *MCPSession) SessionId() string {
	return m.Id
}
