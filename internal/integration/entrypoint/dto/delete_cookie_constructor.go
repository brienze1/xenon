package dto

func NewDeleteCookie(localId, domain, name, path string) DeleteCookie {
	return DeleteCookie{LocalId: localId, Domain: domain, Name: name, Path: path}
}
