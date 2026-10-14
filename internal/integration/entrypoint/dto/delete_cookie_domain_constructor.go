package dto

func NewDeleteCookieDomain(localId, domain string) DeleteCookieDomain {
	return DeleteCookieDomain{LocalId: localId, Domain: domain}
}
