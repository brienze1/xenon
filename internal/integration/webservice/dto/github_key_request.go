package dto

type GitHubKeyRequest struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}

func NewGitHubKeyRequest(title, key string) GitHubKeyRequest {
	return GitHubKeyRequest{Title: title, Key: key}
}
