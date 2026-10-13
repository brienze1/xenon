package dto

type ResetVault struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Confirm bool   `json:"confirm"`
}
