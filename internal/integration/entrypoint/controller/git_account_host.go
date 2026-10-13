package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type gitAccountHost struct {
	errorHandler         adapter.ErrorHandler
	verifyGitHostService appAdapter.VerifyGitHostService
}

func NewGitAccountHost(errorHandler adapter.ErrorHandler, verifyGitHostService appAdapter.VerifyGitHostService) adapter.GitAccountHostController {
	return &gitAccountHost{errorHandler: errorHandler, verifyGitHostService: verifyGitHostService}
}

func (g *gitAccountHost) TestConnection(bindingCtx context.Context, request dto.TestGitConnection) (response *dto.GitRemoteInfo, err error) {
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

	err = validator.CustomValidator().Struct("ACC-59400", &request)
	if err != nil {
		return nil, err
	}

	info, err := g.verifyGitHostService.TestConnection(ctx, request.AccountID, request.Repository)
	if err != nil {
		return nil, err
	}
	result := dto.NewGitRemoteInfo(*info)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (g *gitAccountHost) ConfirmHostKey(bindingCtx context.Context, request dto.ConfirmGitHostKey) (err error) {
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

	err = validator.CustomValidator().Struct("ACC-60400", &request)
	if err != nil {
		return err
	}

	err = g.verifyGitHostService.ConfirmHostKey(ctx, request.Repository, request.Fingerprint, request.Accept)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
