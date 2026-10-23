package dto

type WorkspaceReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}
