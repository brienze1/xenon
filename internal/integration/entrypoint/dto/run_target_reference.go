package dto

type RunTargetReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	ItemId  string `json:"item_id" validate:"required,not_blank"`
}
