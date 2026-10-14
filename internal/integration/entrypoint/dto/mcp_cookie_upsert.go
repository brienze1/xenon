package dto

type MCPCookieUpsert struct {
	Domain string `json:"domain"`
	Raw    string `json:"raw"`
}

func (m *MCPCookieUpsert) UpsertCookie(localId string) UpsertCookie {
	return NewUpsertCookie(localId, m.Domain, m.Raw)
}

func (m MCPCookieUpsert) LogValue() any {
	return map[string]any{"domain": m.Domain, "raw_length": len(m.Raw)}
}
