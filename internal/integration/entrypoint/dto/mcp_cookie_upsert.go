package dto

type MCPCookieUpsert struct {
	Domain string `json:"domain" description:"Domain the cookie belongs to." required_for:"upsert"`
	Raw    string `json:"raw" description:"The cookie as a Set-Cookie header value." required_for:"upsert"`
}

func (m *MCPCookieUpsert) UpsertCookie(localId string) UpsertCookie {
	return NewUpsertCookie(localId, m.Domain, m.Raw)
}

func (m MCPCookieUpsert) LogValue() any {
	return map[string]any{"domain": m.Domain, "raw_length": len(m.Raw)}
}
