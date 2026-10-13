package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/gitaccountkind"
	"github.com/brienze1/xenon/internal/domain/enums/gitprovider"
	"github.com/brienze1/xenon/internal/domain/enums/sshmode"
)

type GitAccount struct {
	CreatedAt     time.Time `json:"createdAt"`
	ID            string    `json:"id"`
	Kind          string    `json:"kind"`
	Provider      string    `json:"provider"`
	SshMode       string    `json:"sshMode,omitempty"`
	DisplayName   string    `json:"displayName"`
	Host          string    `json:"host"`
	Username      string    `json:"username,omitempty"`
	CommitName    string    `json:"commitName"`
	CommitEmail   string    `json:"commitEmail"`
	AvatarURL     string    `json:"avatarUrl,omitempty"`
	KeyPath       string    `json:"keyPath,omitempty"`
	PublicKey     string    `json:"publicKey,omitempty"`
	HasPassphrase bool      `json:"hasPassphrase,omitempty"`
}

func NewGitAccount(account entity.GitAccount) GitAccount {
	return GitAccount{
		CreatedAt: account.CreatedAt.UTC(), ID: account.ID, Kind: account.Kind.String(), Provider: account.Provider.String(),
		SshMode: account.SshMode.String(), DisplayName: account.DisplayName, Host: account.Host, Username: account.Username,
		CommitName: account.CommitName, CommitEmail: account.CommitEmail, AvatarURL: account.AvatarURL, KeyPath: account.KeyPath,
		PublicKey: account.PublicKey, HasPassphrase: account.HasPassphrase,
	}
}

func (g *GitAccount) ToEntity() entity.GitAccount {
	account := entity.GitAccount{
		CreatedAt: g.CreatedAt, ID: g.ID, Kind: gitaccountkind.GetGitAccountKind(g.Kind), Provider: gitprovider.GetGitProvider(g.Provider),
		DisplayName: g.DisplayName, Host: g.Host, Username: g.Username, CommitName: g.CommitName, CommitEmail: g.CommitEmail,
		AvatarURL: g.AvatarURL, KeyPath: g.KeyPath, PublicKey: g.PublicKey, HasPassphrase: g.HasPassphrase,
	}
	if g.SshMode != "" {
		account.SshMode = sshmode.GetSshMode(g.SshMode)
	}
	return account
}
