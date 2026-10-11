package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type gitAccountManage struct {
	errorHandler             adapter.ErrorHandler
	manageGitAccountsService appAdapter.ManageGitAccountsService
}

func NewGitAccountManage(errorHandler adapter.ErrorHandler, manageGitAccountsService appAdapter.ManageGitAccountsService) adapter.GitAccountManageController {
	return &gitAccountManage{errorHandler: errorHandler, manageGitAccountsService: manageGitAccountsService}
}

func (g *gitAccountManage) List(bindingCtx context.Context) (response []dto.GitAccount, err error) {
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

	accounts, err := g.manageGitAccountsService.List(ctx)
	if err != nil {
		return nil, err
	}
	result := dto.NewGitAccounts(accounts)

	logger.Info(ctx, "Finished", map[string]any{"accounts": len(result)})
	return result, nil
}

func (g *gitAccountManage) Update(bindingCtx context.Context, request dto.UpdateGitAccount) (response *dto.GitAccount, err error) {
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

	err = validator.CustomValidator().Struct("ACC-55400", &request)
	if err != nil {
		return nil, err
	}

	account, err := g.manageGitAccountsService.Update(ctx, request.ID, request.DisplayName, request.CommitName, request.CommitEmail)
	if err != nil {
		return nil, err
	}
	result := dto.NewGitAccount(*account)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (g *gitAccountManage) Remove(bindingCtx context.Context, request dto.GitAccountReference) (err error) {
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

	err = validator.CustomValidator().Struct("ACC-56400", &request)
	if err != nil {
		return err
	}

	err = g.manageGitAccountsService.Remove(ctx, request.ID)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
