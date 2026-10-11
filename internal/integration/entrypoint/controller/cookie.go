package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type cookie struct {
	errorHandler         adapter.ErrorHandler
	manageCookiesService appAdapter.ManageCookiesService
}

func NewCookie(errorHandler adapter.ErrorHandler, manageCookiesService appAdapter.ManageCookiesService) adapter.CookieController {
	return &cookie{errorHandler: errorHandler, manageCookiesService: manageCookiesService}
}

func (c *cookie) List(bindingCtx context.Context, request dto.CookieJarReference) (response dto.CookieDomainList, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CKE-01400", &request)
	if err != nil {
		return nil, err
	}

	domains, err := c.manageCookiesService.List(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewCookieDomainList(domains)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (c *cookie) Upsert(bindingCtx context.Context, request dto.UpsertCookie) (response *dto.Cookie, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CKE-02400", &request)
	if err != nil {
		return nil, err
	}

	stored, err := c.manageCookiesService.Upsert(ctx, request.LocalId, request.Domain, request.Raw, request.ReplacedCookie())
	if err != nil {
		return nil, err
	}
	result := dto.NewCookie(*stored)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (c *cookie) Delete(bindingCtx context.Context, request dto.DeleteCookie) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CKE-03400", &request)
	if err != nil {
		return err
	}

	err = c.manageCookiesService.Delete(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (c *cookie) DeleteDomain(bindingCtx context.Context, request dto.DeleteCookieDomain) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CKE-04400", &request)
	if err != nil {
		return err
	}

	err = c.manageCookiesService.DeleteDomain(ctx, request.LocalId, request.Domain)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (c *cookie) Clear(bindingCtx context.Context, request dto.CookieJarReference) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CKE-05400", &request)
	if err != nil {
		return err
	}

	err = c.manageCookiesService.Clear(ctx, request.LocalId)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (c *cookie) GetAllowlist(bindingCtx context.Context, request dto.CookieJarReference) (response *dto.CookieAllowlist, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CKE-06400", &request)
	if err != nil {
		return nil, err
	}

	allowlist, err := c.manageCookiesService.GetAllowlist(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewCookieAllowlist(*allowlist)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (c *cookie) SaveAllowlist(bindingCtx context.Context, request dto.SaveCookieAllowlist) (response *dto.CookieAllowlist, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, c.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("CKE-07400", &request)
	if err != nil {
		return nil, err
	}

	allowlist, err := c.manageCookiesService.SaveAllowlist(ctx, request.LocalId, request.Domains)
	if err != nil {
		return nil, err
	}
	result := dto.NewCookieAllowlist(*allowlist)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
