package dto

type MoveItem struct {
	LocalId     string `json:"local_id" validate:"required,not_blank"`
	Id          string `json:"id" validate:"required,not_blank"`
	NewParentId string `json:"new_parent_id"`
	Index       int    `json:"index" validate:"min=0"`
}
