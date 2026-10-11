package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type gitAccountAdd struct {
	errorHandler         adapter.ErrorHandler
	addGitAccountService appAdapter.AddGitAccountService
}

func NewGitAccountAdd(errorHandler adapter.ErrorHandler, addGitAccountService appAdapter.AddGitAccountService) adapter.GitAccountAddController {
	return &gitAccountAdd{errorHandler: errorHandler, addGitAccountService: addGitAccountService}
}

func (g *gitAccountAdd) AddToken(bindingCtx context.Context, request dto.AddGitToken) (response *dto.GitAccount, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, g.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("ACC-20400", &request)
	if err != nil {
		return nil, err
	}

	account, err := g.addGitAccountService.AddToken(ctx, request.ToEntity(), request.Token, request.Repository)
	if err != nil {
		return nil, err
	}
	result := dto.NewGitAccount(*account)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (g *gitAccountAdd) AddSSH(bindingCtx context.Context, request dto.AddGitSsh) (response *dto.GitSshAccount, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, g.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("ACC-54400", &request)
	if err != nil {
		return nil, err
	}

	account, publicKey, err := g.addGitAccountService.AddSsh(ctx, request.ToEntity(), request.KeyPath, request.Passphrase)
	if err != nil {
		return nil, err
	}
	result := dto.NewGitSshAccount(*account, publicKey)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
