package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type GitSshAccount struct {
	PublicKey string     `json:"public_key"`
	Account   GitAccount `json:"account"`
}

func NewGitSshAccount(account entity.GitAccount, publicKey string) GitSshAccount {
	return GitSshAccount{Account: NewGitAccount(account), PublicKey: publicKey}
}
