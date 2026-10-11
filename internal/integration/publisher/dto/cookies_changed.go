package dto

const CookiesChangedEvent = "cookies:changed"

type CookiesChanged struct {
	LocalId string `json:"local_id"`
}

func NewCookiesChanged(localId string) CookiesChanged {
	return CookiesChanged{LocalId: localId}
}
