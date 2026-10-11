package dto

type OAuthTokenUse struct {
	LocalId  string `json:"local_id" validate:"required,not_blank"`
	OwnerId  string `json:"owner_id" validate:"required,not_blank"`
	UseToken string `json:"use_token" validate:"required,oneof=access_token id_token"`
}
