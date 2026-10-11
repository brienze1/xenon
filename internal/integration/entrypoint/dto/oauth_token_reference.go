package dto

type OAuthTokenReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	OwnerId string `json:"owner_id" validate:"required,not_blank"`
}
