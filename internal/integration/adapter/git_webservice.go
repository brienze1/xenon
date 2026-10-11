package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type GitWebservice interface {
	usecase.ListRemote
	usecase.VerifyGitToken
	usecase.ScanHostKey
	usecase.CloneRepository
	usecase.FetchBranch
	usecase.PushBranch
	usecase.ReadTreeVersions
	usecase.WriteSyncCommit
	usecase.CheckoutBranch
}
