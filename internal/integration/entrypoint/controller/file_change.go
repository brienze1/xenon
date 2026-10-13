package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type fileChange struct {
	errorHandler            adapter.ErrorHandler
	handleFileChangeService appAdapter.HandleFileChangeService
}

func NewFileChange(errorHandler adapter.ErrorHandler, handleFileChangeService appAdapter.HandleFileChangeService) adapter.FileChangeController {
	return &fileChange{errorHandler: errorHandler, handleFileChangeService: handleFileChangeService}
}

func (f *fileChange) Handle(eventCtx context.Context, request dto.FileChange) (err error) {
	ctx := logger.GetContext(eventCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = f.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("WKS-28400", &request)
	if err != nil {
		return err
	}

	err = f.handleFileChangeService.Handle(ctx, request.LocalId, request.Path)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
