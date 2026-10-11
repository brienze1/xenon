package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type gitAccountGitHub struct {
	errorHandler        adapter.ErrorHandler
	signInGitHubService appAdapter.SignInGitHubService
}

func NewGitAccountGitHub(errorHandler adapter.ErrorHandler, signInGitHubService appAdapter.SignInGitHubService) adapter.GitAccountGitHubController {
	return &gitAccountGitHub{errorHandler: errorHandler, signInGitHubService: signInGitHubService}
}

func (g *gitAccountGitHub) StartGitHubSignIn(bindingCtx context.Context) (response *dto.GitHubDeviceCode, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, g.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, nil)

	code, err := g.signInGitHubService.Start(ctx)
	if err != nil {
		return nil, err
	}
	result := dto.NewGitHubDeviceCode(*code)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (g *gitAccountGitHub) PollGitHubSignIn(bindingCtx context.Context, request dto.GitHubSignInSession) (response *dto.GitHubSignInResult, err error) {
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

	err = validator.CustomValidator().Struct("ACC-57400", &request)
	if err != nil {
		return nil, err
	}

	signIn, account, err := g.signInGitHubService.Poll(ctx, request.SessionID)
	if err != nil {
		return nil, err
	}
	result := dto.NewGitHubSignInResult(*signIn, account)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (g *gitAccountGitHub) UploadPublicKey(bindingCtx context.Context, request dto.UploadGitPublicKey) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = g.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("ACC-58400", &request)
	if err != nil {
		return err
	}

	err = g.signInGitHubService.UploadPublicKey(ctx, request.SshAccountID, request.GitHubAccountID, request.Title)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
