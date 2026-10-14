package dto

type MCPCookieAllowlist struct {
	Domains []string `json:"domains"`
}

func (m *MCPCookieAllowlist) SaveCookieAllowlist(localId string) SaveCookieAllowlist {
	return NewSaveCookieAllowlist(localId, m.Domains)
}
