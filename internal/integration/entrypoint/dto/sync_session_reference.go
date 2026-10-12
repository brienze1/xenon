package dto

type SyncSessionReference struct {
	SessionId string `json:"session_id" validate:"required,not_blank"`
}
