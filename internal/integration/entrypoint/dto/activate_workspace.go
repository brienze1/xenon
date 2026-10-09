package dto

type ActivateWorkspace struct {
	LocalId string `json:"local_id" validate:"required"`
}
