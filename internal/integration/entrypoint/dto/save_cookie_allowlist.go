package dto

type SaveCookieAllowlist struct {
	LocalId string   `json:"local_id" validate:"required,not_blank"`
	Domains []string `json:"domains"`
}
