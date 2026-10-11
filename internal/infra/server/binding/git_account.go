package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type GitAccountService struct {
	manage adapter.GitAccountManageController
	add    adapter.GitAccountAddController
	github adapter.GitAccountGitHubController
	host   adapter.GitAccountHostController
}

func NewGitAccountService(
	manage adapter.GitAccountManageController,
	add adapter.GitAccountAddController,
	github adapter.GitAccountGitHubController,
	host adapter.GitAccountHostController,
) *GitAccountService {
	return &GitAccountService{manage: manage, add: add, github: github, host: host}
}

func (g *GitAccountService) List(ctx context.Context) ([]dto.GitAccount, error) {
	return g.manage.List(ctx)
}

func (g *GitAccountService) Update(ctx context.Context, request dto.UpdateGitAccount) (*dto.GitAccount, error) {
	return g.manage.Update(ctx, request)
}

func (g *GitAccountService) Remove(ctx context.Context, request dto.GitAccountReference) error {
	return g.manage.Remove(ctx, request)
}

func (g *GitAccountService) AddToken(ctx context.Context, request dto.AddGitToken) (*dto.GitAccount, error) {
	return g.add.AddToken(ctx, request)
}

func (g *GitAccountService) AddSSH(ctx context.Context, request dto.AddGitSsh) (*dto.GitSshAccount, error) {
	return g.add.AddSSH(ctx, request)
}

func (g *GitAccountService) StartGitHubSignIn(ctx context.Context) (*dto.GitHubDeviceCode, error) {
	return g.github.StartGitHubSignIn(ctx)
}

func (g *GitAccountService) PollGitHubSignIn(ctx context.Context, request dto.GitHubSignInSession) (*dto.GitHubSignInResult, error) {
	return g.github.PollGitHubSignIn(ctx, request)
}

func (g *GitAccountService) UploadPublicKey(ctx context.Context, request dto.UploadGitPublicKey) error {
	return g.github.UploadPublicKey(ctx, request)
}

func (g *GitAccountService) TestConnection(ctx context.Context, request dto.TestGitConnection) (*dto.GitRemoteInfo, error) {
	return g.host.TestConnection(ctx, request)
}

func (g *GitAccountService) ConfirmHostKey(ctx context.Context, request dto.ConfirmGitHostKey) error {
	return g.host.ConfirmHostKey(ctx, request)
}
