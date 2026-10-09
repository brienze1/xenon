package dto

type RemoveWorkspace struct {
	LocalId     string `json:"local_id" validate:"required,not_blank"`
	MoveToTrash bool   `json:"move_to_trash"`
}
