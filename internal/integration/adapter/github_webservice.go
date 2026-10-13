package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type GitHubWebservice interface {
	usecase.SignInWithGitHub
	usecase.FindGitHubUser
	usecase.UploadSshPublicKey
}
