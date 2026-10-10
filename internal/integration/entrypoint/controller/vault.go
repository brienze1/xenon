package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type vault struct {
	errorHandler       adapter.ErrorHandler
	manageVaultService appAdapter.ManageVaultService
}

func NewVault(errorHandler adapter.ErrorHandler, manageVaultService appAdapter.ManageVaultService) adapter.VaultController {
	return &vault{errorHandler: errorHandler, manageVaultService: manageVaultService}
}

func (v *vault) Status(bindingCtx context.Context) (response *dto.VaultStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, nil)

	status, err := v.manageVaultService.Status(ctx)
	if err != nil {
		return nil, err
	}
	result := dto.NewVaultStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *vault) SetupPassphrase(bindingCtx context.Context, request dto.SetupPassphrase) (response *dto.VaultStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VLT-10400", &request)
	if err != nil {
		return nil, err
	}

	status, err := v.manageVaultService.SetupPassphrase(ctx, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewVaultStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *vault) Unlock(bindingCtx context.Context, request dto.UnlockVault) (response *dto.VaultStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VLT-11400", &request)
	if err != nil {
		return nil, err
	}

	status, err := v.manageVaultService.Unlock(ctx, request.Passphrase)
	if err != nil {
		return nil, err
	}
	result := dto.NewVaultStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *vault) Reveal(bindingCtx context.Context, request dto.SecretReference) (response *dto.RevealedSecret, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VLT-12400", &request)
	if err != nil {
		return nil, err
	}

	value, err := v.manageVaultService.Reveal(ctx, request.LocalId, request.ScopeRef(), request.Key)
	if err != nil {
		return nil, err
	}
	result := dto.NewRevealedSecret(value)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (v *vault) CopySecret(bindingCtx context.Context, request dto.SecretReference) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VLT-13400", &request)
	if err != nil {
		return err
	}

	if err := v.manageVaultService.CopySecret(ctx, request.LocalId, request.ScopeRef(), request.Key); err != nil {
		return err
	}

	logger.Info(ctx, "Finished")
	return nil
}

func (v *vault) Reset(bindingCtx context.Context, request dto.ResetVault) (response *dto.VaultStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, v.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("VLT-14400", &request)
	if err != nil {
		return nil, err
	}

	status, err := v.manageVaultService.Reset(ctx, request.LocalId, request.Confirm)
	if err != nil {
		return nil, err
	}
	result := dto.NewVaultStatus(*status)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
