package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type treeSearch struct {
	errorHandler      adapter.ErrorHandler
	searchTreeService appAdapter.SearchTreeService
}

func NewTreeSearch(errorHandler adapter.ErrorHandler, searchTreeService appAdapter.SearchTreeService) adapter.TreeSearchController {
	return &treeSearch{errorHandler: errorHandler, searchTreeService: searchTreeService}
}

func (t *treeSearch) GetTree(bindingCtx context.Context, request dto.GetTree) (response dto.TreeNodeList, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("TRE-04400", &request)
	if err != nil {
		return nil, err
	}

	tree, err := t.searchTreeService.Tree(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeNodeList(*tree)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (t *treeSearch) Search(bindingCtx context.Context, request dto.SearchTree) (response *dto.TreeSearch, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, t.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validator.CustomValidator().Struct("TRE-05400", &request)
	if err != nil {
		return nil, err
	}

	search, err := t.searchTreeService.Search(ctx, request.LocalId, request.Query)
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeSearch(*search)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
