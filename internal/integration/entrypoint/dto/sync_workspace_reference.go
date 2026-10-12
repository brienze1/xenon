package dto

type SyncWorkspaceReference struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
}
