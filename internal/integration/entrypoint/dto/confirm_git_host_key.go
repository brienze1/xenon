package dto

type ConfirmGitHostKey struct {
	Repository  string `json:"repository" validate:"required,not_blank"`
	Fingerprint string `json:"fingerprint"`
	Accept      bool   `json:"accept"`
}
