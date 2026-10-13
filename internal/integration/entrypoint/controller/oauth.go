package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type oauth struct {
	errorHandler       adapter.ErrorHandler
	manageOAuthService appAdapter.ManageOAuthService
}

func NewOAuth(errorHandler adapter.ErrorHandler, manageOAuthService appAdapter.ManageOAuthService) adapter.OAuthController {
	return &oauth{errorHandler: errorHandler, manageOAuthService: manageOAuthService}
}

func (o *oauth) GetNewToken(bindingCtx context.Context, request dto.OAuthTokenRequest) (response *dto.TokenInfo, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, o.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("OAU-14400", &request)
	if err != nil {
		return nil, err
	}

	token, err := o.manageOAuthService.GetNewToken(ctx, request.LocalId, request.OwnerId, request.Draft())
	if err != nil {
		return nil, err
	}
	result := dto.NewTokenInfo(*token, timeutils.Now())

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (o *oauth) Refresh(bindingCtx context.Context, request dto.OAuthTokenRequest) (response *dto.TokenInfo, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, o.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("OAU-15400", &request)
	if err != nil {
		return nil, err
	}

	token, err := o.manageOAuthService.Refresh(ctx, request.LocalId, request.OwnerId, request.Draft())
	if err != nil {
		return nil, err
	}
	result := dto.NewTokenInfo(*token, timeutils.Now())

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (o *oauth) Get(bindingCtx context.Context, request dto.OAuthTokenReference) (response *dto.TokenInfo, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, o.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("OAU-16400", &request)
	if err != nil {
		return nil, err
	}

	token, err := o.manageOAuthService.Get(ctx, request.LocalId, request.OwnerId)
	if err != nil {
		return nil, err
	}
	if token == nil {
		logger.Info(ctx, "Finished", map[string]any{"found": false})
		return nil, nil
	}
	result := dto.NewTokenInfo(*token, timeutils.Now())

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (o *oauth) Reveal(bindingCtx context.Context, request dto.OAuthTokenPart) (response *dto.RevealedSecret, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, o.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("OAU-17400", &request)
	if err != nil {
		return nil, err
	}

	value, err := o.manageOAuthService.Reveal(ctx, request.LocalId, request.OwnerId, request.Which)
	if err != nil {
		return nil, err
	}
	result := dto.NewRevealedSecret(value)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (o *oauth) Copy(bindingCtx context.Context, request dto.OAuthTokenPart) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = o.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("OAU-18400", &request)
	if err != nil {
		return err
	}

	err = o.manageOAuthService.Copy(ctx, request.LocalId, request.OwnerId, request.Which)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (o *oauth) Use(bindingCtx context.Context, request dto.OAuthTokenUse) (response *dto.TokenInfo, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, o.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("OAU-19400", &request)
	if err != nil {
		return nil, err
	}

	token, err := o.manageOAuthService.Use(ctx, request.LocalId, request.OwnerId, request.UseToken)
	if err != nil {
		return nil, err
	}
	result := dto.NewTokenInfo(*token, timeutils.Now())

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (o *oauth) Clear(bindingCtx context.Context, request dto.OAuthTokenReference) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = o.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("OAU-20400", &request)
	if err != nil {
		return err
	}

	err = o.manageOAuthService.Clear(ctx, request.LocalId, request.OwnerId)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
