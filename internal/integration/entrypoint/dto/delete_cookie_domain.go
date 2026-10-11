package dto

type DeleteCookieDomain struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Domain  string `json:"domain" validate:"required,not_blank"`
}
