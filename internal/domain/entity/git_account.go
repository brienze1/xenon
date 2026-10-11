package entity

import (
	"net"
	"strings"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/gitaccountkind"
	"github.com/brienze1/xenon/internal/domain/enums/gitprovider"
	"github.com/brienze1/xenon/internal/domain/enums/sshmode"
)

const (
	gitHubHost                  = "github.com"
	gitHubTokenUsername         = "x-access-token"
	gitLabTokenUsername         = "oauth2"
	bitbucketApiTokenUsername   = "x-bitbucket-api-token-auth"
	accountDisplayNameSeparator = " @ "
)

type GitAccount struct {
	CreatedAt     time.Time
	ID            string
	Kind          enums.GitAccountKind
	Provider      enums.GitProvider
	SshMode       enums.SshMode
	DisplayName   string
	Host          string
	Username      string
	CommitName    string
	CommitEmail   string
	AvatarURL     string
	KeyPath       string
	PublicKey     string
	HasPassphrase bool
}

func NewGitHubAccount(id string, user GitHubUser, createdAt time.Time) GitAccount {
	return GitAccount{
		ID:          id,
		Kind:        gitaccountkind.GitHub,
		Provider:    gitprovider.GitHub,
		DisplayName: user.Login + accountDisplayNameSeparator + gitHubHost,
		Host:        gitHubHost,
		Username:    user.Login,
		CommitName:  user.CommitName(),
		CommitEmail: user.CommitEmail(),
		AvatarURL:   user.AvatarURL,
		CreatedAt:   createdAt,
	}
}

func NewTokenAccount(id string, provider enums.GitProvider, host, username string, createdAt time.Time) GitAccount {
	account := GitAccount{
		ID:        id,
		Kind:      gitaccountkind.Token,
		Provider:  provider,
		Host:      strings.ToLower(strings.TrimSpace(host)),
		Username:  strings.TrimSpace(username),
		CreatedAt: createdAt,
	}
	account.DisplayName = account.defaultDisplayName()
	return account
}

func NewSshAccount(id string, mode enums.SshMode, host string, key SshKey, hasPassphrase bool, createdAt time.Time) GitAccount {
	account := GitAccount{
		ID:            id,
		Kind:          gitaccountkind.Ssh,
		Provider:      gitprovider.Other,
		SshMode:       mode,
		Host:          strings.ToLower(strings.TrimSpace(host)),
		Username:      "git",
		KeyPath:       key.Path,
		PublicKey:     key.PublicKey,
		HasPassphrase: hasPassphrase && mode == sshmode.KeyFile,
		CreatedAt:     createdAt,
	}
	account.DisplayName = "SSH" + accountDisplayNameSeparator + account.Host
	return account
}

func (g GitAccount) WithIdentity(displayName, commitName, commitEmail string) GitAccount {
	if strings.TrimSpace(displayName) != "" {
		g.DisplayName = strings.TrimSpace(displayName)
	}
	g.CommitName = strings.TrimSpace(commitName)
	g.CommitEmail = strings.TrimSpace(commitEmail)
	return g
}

func (g GitAccount) WithGitHubUser(user GitHubUser) GitAccount {
	g.Username = user.Login
	g.AvatarURL = user.AvatarURL
	if g.CommitName == "" {
		g.CommitName = user.CommitName()
	}
	if g.CommitEmail == "" {
		g.CommitEmail = user.CommitEmail()
	}
	g.DisplayName = g.defaultDisplayName()
	return g
}

func (g GitAccount) IsSsh() bool {
	return g.Kind == gitaccountkind.Ssh
}

func (g GitAccount) UsesAgent() bool {
	return g.IsSsh() && g.SshMode == sshmode.Agent
}

func (g GitAccount) IsGenerated() bool {
	return g.IsSsh() && g.SshMode == sshmode.Generated
}

func (g GitAccount) CanCallGitHub() bool {
	return g.Kind == gitaccountkind.GitHub || (g.Kind == gitaccountkind.Token && g.Provider == gitprovider.GitHub)
}

func (g GitAccount) StoresSecret() bool {
	return !g.IsSsh() || g.HasPassphrase
}

func (g GitAccount) BasicAuthUsername() string {
	switch g.Provider {
	case gitprovider.GitHub:
		if g.Username == "" {
			return gitHubTokenUsername
		}
		return g.Username
	case gitprovider.GitLab:
		return gitLabTokenUsername
	case gitprovider.Bitbucket:
		return bitbucketApiTokenUsername
	case gitprovider.AzureDevOps:
		return ""
	default:
		return g.Username
	}
}

func (g GitAccount) ServesHost(host string) bool {
	wanted := strings.ToLower(host)
	if g.Host == wanted {
		return true
	}
	hostname, _, err := net.SplitHostPort(wanted)
	return err == nil && g.Host == hostname
}

func (g GitAccount) defaultDisplayName() string {
	if g.Username == "" {
		return g.Host
	}
	return g.Username + accountDisplayNameSeparator + g.Host
}
