package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type oauthTokens struct {
	vaults            vaultWriter
	lockOAuthToken    usecase.LockOAuthToken
	refreshOAuthToken usecase.RefreshOAuthToken
}

func (o oauthTokens) find(ctx context.Context, localId, owner string) (*entity.OAuth2Token, error) {
	vault, err := o.vaults.findVault.Find(ctx, localId)
	if err != nil {
		return nil, err
	}
	token, found := vault.TokenOf(owner)
	if !found {
		return nil, nil
	}
	return &token, nil
}

func (o oauthTokens) save(ctx context.Context, localId string, token entity.OAuth2Token) error {
	_, err := o.vaults.update(ctx, localId, func(vault entity.Vault) entity.Vault {
		return vault.WithToken(token)
	})
	return err
}

func (o oauthTokens) delete(ctx context.Context, localId, owner string) error {
	token, err := o.find(ctx, localId, owner)
	if err != nil || token == nil {
		return err
	}
	_, err = o.vaults.update(ctx, localId, func(vault entity.Vault) entity.Vault {
		return vault.WithoutToken(owner)
	})
	return err
}

func (o oauthTokens) current(ctx context.Context, localId, owner string, config entity.OAuth2Config) (*entity.OAuth2Token, bool, error) {
	unlock := o.lockOAuthToken.LockToken(ctx, localId, owner)
	defer unlock()
	token, err := o.find(ctx, localId, owner)
	if err != nil || token == nil {
		return token, false, err
	}
	if !token.IsExpired(timeutils.Now()) {
		return token, false, nil
	}
	if !config.AutoRefresh || !token.CanRefresh() {
		return token, true, nil
	}
	fresh, err := o.renew(ctx, localId, config, *token)
	if err != nil {
		logger.Warn(ctx, err, "The expired OAuth 2.0 token could not be refreshed", map[string]any{"owner": owner})
		return token, true, nil
	}
	return fresh, false, nil
}

func (o oauthTokens) refresh(ctx context.Context, localId, owner string, config entity.OAuth2Config) (*entity.OAuth2Token, error) {
	unlock := o.lockOAuthToken.LockToken(ctx, localId, owner)
	defer unlock()
	token, err := o.find(ctx, localId, owner)
	if err != nil {
		return nil, err
	}
	if token == nil || !token.CanRefresh() {
		return nil, usecase.RejectMissingRefreshToken()
	}
	return o.renew(ctx, localId, config, *token)
}

func (o oauthTokens) renew(ctx context.Context, localId string, config entity.OAuth2Config, token entity.OAuth2Token) (*entity.OAuth2Token, error) {
	fresh, err := o.refreshOAuthToken.Refresh(ctx, config, token)
	if err != nil {
		return nil, err
	}
	refreshed := token.Refreshed(*fresh)
	if err := o.save(ctx, localId, refreshed); err != nil {
		return nil, err
	}
	return &refreshed, nil
}
