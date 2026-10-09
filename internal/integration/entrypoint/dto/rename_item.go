package dto

type RenameItem struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	Id      string `json:"id" validate:"required,not_blank"`
	Name    string `json:"name" validate:"required,not_blank,max=200"`
}
