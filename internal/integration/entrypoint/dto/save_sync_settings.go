package dto

type SaveSyncSettings struct {
	LocalId   string `json:"local_id" validate:"required,not_blank"`
	AccountID string `json:"account_id"`
	AutoPull  bool   `json:"auto_pull"`
	AutoPush  bool   `json:"auto_push"`
}
