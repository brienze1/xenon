package dto

type GitHubSignInSession struct {
	SessionID string `json:"session_id" validate:"required,not_blank"`
}
