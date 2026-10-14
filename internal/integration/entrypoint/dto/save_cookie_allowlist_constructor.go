package dto

func NewSaveCookieAllowlist(localId string, domains []string) SaveCookieAllowlist {
	return SaveCookieAllowlist{LocalId: localId, Domains: domains}
}
