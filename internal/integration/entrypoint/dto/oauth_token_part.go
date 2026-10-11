package dto

type OAuthTokenPart struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	OwnerId string `json:"owner_id" validate:"required,not_blank"`
	Which   string `json:"which" validate:"required,oneof=access refresh id"`
}
