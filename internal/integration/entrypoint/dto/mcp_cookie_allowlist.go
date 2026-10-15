package dto

type MCPCookieAllowlist struct {
	Domains []string `json:"domains" description:"Domains scripts may read cookies from; replaces the allowlist."`
}

func (m *MCPCookieAllowlist) SaveCookieAllowlist(localId string) SaveCookieAllowlist {
	return NewSaveCookieAllowlist(localId, m.Domains)
}
