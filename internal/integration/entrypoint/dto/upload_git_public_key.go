package dto

type UploadGitPublicKey struct {
	SshAccountID    string `json:"ssh_account_id" validate:"required,not_blank"`
	GitHubAccountID string `json:"github_account_id" validate:"required,not_blank"`
	Title           string `json:"title" validate:"required,not_blank"`
}
