package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type manageOAuth struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	authorizeInBrowser    usecase.AuthorizeInBrowser
	exchangeOAuthToken    usecase.ExchangeOAuthToken
	copyToClipboard       usecase.CopyToClipboard
	owners                oauthOwners
	tokens                oauthTokens
}

func NewManageOAuthService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	findRequest usecase.FindRequest,
	loadWorkspace usecase.LoadWorkspace,
	findEnvironment usecase.FindEnvironment,
	findLocalVariables usecase.FindLocalVariables,
	findVault usecase.FindVault,
	saveVault usecase.SaveVault,
	findVaultSettings usecase.FindVaultSettings,
	saveVaultSettings usecase.SaveVaultSettings,
	probeKeychain usecase.ProbeKeychain,
	publishVariableEvent usecase.PublishVariableEvent,
	authorizeInBrowser usecase.AuthorizeInBrowser,
	exchangeOAuthToken usecase.ExchangeOAuthToken,
	refreshOAuthToken usecase.RefreshOAuthToken,
	lockOAuthToken usecase.LockOAuthToken,
	copyToClipboard usecase.CopyToClipboard,
) adapter.ManageOAuthService {
	return &manageOAuth{
		findWorkspaceRegistry: findWorkspaceRegistry,
		authorizeInBrowser:    authorizeInBrowser,
		exchangeOAuthToken:    exchangeOAuthToken,
		copyToClipboard:       copyToClipboard,
		owners: oauthOwners{
			findTreeItem: findTreeItem,
			findRequest:  findRequest,
			variables: requestVariables{
				findTreeItem:       findTreeItem,
				loadWorkspace:      loadWorkspace,
				findEnvironment:    findEnvironment,
				findLocalVariables: findLocalVariables,
				findVault:          findVault,
			},
		},
		tokens: oauthTokens{
			vaults: vaultWriter{
				findVault:            findVault,
				saveVault:            saveVault,
				findVaultSettings:    findVaultSettings,
				saveVaultSettings:    saveVaultSettings,
				probeKeychain:        probeKeychain,
				publishVariableEvent: publishVariableEvent,
			},
			lockOAuthToken:    lockOAuthToken,
			refreshOAuthToken: refreshOAuthToken,
		},
	}
}

func (m *manageOAuth) GetNewToken(ctx context.Context, localId, owner string, draft *entity.OAuth2Config) (*entity.OAuth2Token, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "owner": owner, "draft": draft})

	config, err := m.config(ctx, localId, owner, draft)
	if err != nil {
		return nil, err
	}
	if !config.CanObtainToken() {
		return nil, usecase.RejectInvalidOAuthConfig(config.InvalidReason())
	}
	token, err := m.obtain(ctx, *config)
	if err != nil {
		return nil, err
	}
	owned := token.OwnedBy(owner, config.UseToken)
	if err := m.tokens.save(ctx, localId, owned); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "result": owned})
	return &owned, nil
}

func (m *manageOAuth) Refresh(ctx context.Context, localId, owner string, draft *entity.OAuth2Config) (*entity.OAuth2Token, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "owner": owner, "draft": draft})

	config, err := m.config(ctx, localId, owner, draft)
	if err != nil {
		return nil, err
	}
	token, err := m.tokens.refresh(ctx, localId, owner, *config)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "result": token})
	return token, nil
}

func (m *manageOAuth) Get(ctx context.Context, localId, owner string) (*entity.OAuth2Token, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "owner": owner})

	token, err := m.tokens.find(ctx, localId, owner)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "found": token != nil})
	return token, nil
}

func (m *manageOAuth) Reveal(ctx context.Context, localId, owner, which string) (string, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "owner": owner, "which": which})

	value, err := m.part(ctx, localId, owner, which)
	if err != nil {
		return "", err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "which": which})
	return value, nil
}

func (m *manageOAuth) Copy(ctx context.Context, localId, owner, which string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "owner": owner, "which": which})

	value, err := m.part(ctx, localId, owner, which)
	if err != nil {
		return err
	}
	if err := m.copyToClipboard.Copy(ctx, value); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "which": which})
	return nil
}

func (m *manageOAuth) Use(ctx context.Context, localId, owner, useToken string) (*entity.OAuth2Token, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "owner": owner, "useToken": useToken})

	token, err := m.tokens.find(ctx, localId, owner)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, usecase.RejectMissingOAuthToken()
	}
	used := token.Using(useToken)
	if err := m.tokens.save(ctx, localId, used); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "result": used})
	return &used, nil
}

func (m *manageOAuth) Clear(ctx context.Context, localId, owner string) error {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "owner": owner})

	if err := m.tokens.delete(ctx, localId, owner); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "owner": owner})
	return nil
}

func (m *manageOAuth) part(ctx context.Context, localId, owner, which string) (string, error) {
	token, err := m.tokens.find(ctx, localId, owner)
	if err != nil {
		return "", err
	}
	if token == nil {
		return "", usecase.RejectMissingOAuthToken()
	}
	return token.Part(which), nil
}

func (m *manageOAuth) config(ctx context.Context, localId, owner string, draft *entity.OAuth2Config) (*entity.OAuth2Config, error) {
	workspace, err := m.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	return m.owners.config(ctx, localId, workspace.Path, owner, draft)
}

func (m *manageOAuth) obtain(ctx context.Context, config entity.OAuth2Config) (*entity.OAuth2Token, error) {
	grant := entity.NewOAuth2GrantRequest(config)
	if !config.UsesBrowser() {
		return m.exchangeOAuthToken.Exchange(ctx, grant)
	}
	state := usecase.GenerateOAuthState(config)
	pkce := usecase.GeneratePkce(config)
	callback, err := m.authorizeInBrowser.Authorize(ctx, usecase.BuildAuthorizationUrl(config, state, pkce), config.RedirectURI)
	if err != nil {
		return nil, err
	}
	if callback.IsError() {
		return nil, usecase.RejectOAuthProviderError(callback.Error, callback.ErrorDescription)
	}
	if !callback.MatchesState(state) {
		return nil, usecase.RejectOAuthStateMismatch()
	}
	if config.IsImplicit() {
		token := callback.Token(timeutils.Now(), config.Scope)
		return &token, nil
	}
	return m.exchangeOAuthToken.Exchange(ctx, grant.WithCode(callback.Code, pkce.Verifier))
}
