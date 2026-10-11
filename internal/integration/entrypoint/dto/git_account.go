package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type GitAccount struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Provider      string `json:"provider"`
	SshMode       string `json:"ssh_mode"`
	DisplayName   string `json:"display_name"`
	Host          string `json:"host"`
	Username      string `json:"username"`
	CommitName    string `json:"commit_name"`
	CommitEmail   string `json:"commit_email"`
	AvatarURL     string `json:"avatar_url"`
	KeyPath       string `json:"key_path"`
	PublicKey     string `json:"public_key"`
	CreatedAt     string `json:"created_at"`
	HasPassphrase bool   `json:"has_passphrase"`
}

func NewGitAccount(account entity.GitAccount) GitAccount {
	return GitAccount{
		ID: account.ID, Kind: account.Kind.String(), Provider: account.Provider.String(), SshMode: account.SshMode.String(),
		DisplayName: account.DisplayName, Host: account.Host, Username: account.Username, CommitName: account.CommitName,
		CommitEmail: account.CommitEmail, AvatarURL: account.AvatarURL, KeyPath: account.KeyPath, PublicKey: account.PublicKey,
		CreatedAt: account.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"), HasPassphrase: account.HasPassphrase,
	}
}

func NewGitAccounts(accounts []entity.GitAccount) []GitAccount {
	converted := make([]GitAccount, 0, len(accounts))
	for _, account := range accounts {
		converted = append(converted, NewGitAccount(account))
	}
	return converted
}
