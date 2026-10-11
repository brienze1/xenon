package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type GitHubSignInResult struct {
	Account  *GitAccount `json:"account,omitempty"`
	Status   string      `json:"status"`
	Interval int         `json:"interval"`
}

func NewGitHubSignInResult(signIn entity.GitHubSignIn, account *entity.GitAccount) GitHubSignInResult {
	result := GitHubSignInResult{Status: signIn.Status.String(), Interval: signIn.Interval}
	if account != nil {
		converted := NewGitAccount(*account)
		result.Account = &converted
	}
	return result
}
