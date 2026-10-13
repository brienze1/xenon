package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type GitAccountPersistence interface {
	usecase.FindGitAccounts
	usecase.SaveGitAccount
	usecase.DeleteGitAccount
}
