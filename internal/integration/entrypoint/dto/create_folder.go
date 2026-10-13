package dto

type CreateFolder struct {
	LocalId  string `json:"local_id" validate:"required,not_blank"`
	ParentId string `json:"parent_id" validate:"required,not_blank"`
	Name     string `json:"name" validate:"required,not_blank,max=200"`
}
