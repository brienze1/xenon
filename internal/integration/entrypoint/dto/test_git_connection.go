package dto

type TestGitConnection struct {
	AccountID  string `json:"account_id" validate:"required,not_blank"`
	Repository string `json:"repository" validate:"required,not_blank"`
}
