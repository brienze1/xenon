package dto

type CookieJarReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}
