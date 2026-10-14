package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPCookiesList struct {
	Domain string `json:"domain"`
	Cursor string `json:"cursor"`
}

// Page filters the jar by the domain argument and pages it from the cursor; false means an unknown cursor.
func (m *MCPCookiesList) Page(domains []entity.CookieDomain) (entity.AIPage[entity.Cookie], bool) {
	return entity.NewAIPage(entity.CookiesMatching(domains, m.Domain), m.Cursor)
}
