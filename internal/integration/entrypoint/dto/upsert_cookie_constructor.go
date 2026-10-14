package dto

func NewUpsertCookie(localId, domain, raw string) UpsertCookie {
	return UpsertCookie{LocalId: localId, Domain: domain, Raw: raw}
}
