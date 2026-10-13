package dto

type UpdateGitAccount struct {
	ID          string `json:"id" validate:"required,not_blank"`
	DisplayName string `json:"display_name"`
	CommitName  string `json:"commit_name"`
	CommitEmail string `json:"commit_email" validate:"omitempty,email"`
}
