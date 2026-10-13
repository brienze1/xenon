package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type UpsertCookie struct {
	Replaces *CookieKey `json:"replaces,omitempty"`
	LocalId  string     `json:"local_id" validate:"required,not_blank"`
	Domain   string     `json:"domain" validate:"required,not_blank"`
	Raw      string     `json:"raw" validate:"required,not_blank"`
}

func (u UpsertCookie) ReplacedCookie() *entity.Cookie {
	if u.Replaces == nil {
		return nil
	}
	replaced := u.Replaces.ToEntity(u.Domain)
	return &replaced
}

func (u UpsertCookie) LogValue() any {
	return map[string]any{"localId": u.LocalId, "domain": u.Domain, "replaces": u.Replaces}
}
