package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type treeManage struct {
	errorHandler          adapter.ErrorHandler
	manageTreeItemService appAdapter.ManageTreeItemService
}

func NewTreeManage(errorHandler adapter.ErrorHandler, manageTreeItemService appAdapter.ManageTreeItemService) adapter.TreeManageController {
	return &treeManage{errorHandler: errorHandler, manageTreeItemService: manageTreeItemService}
}

func (t *treeManage) CreateCollection(bindingCtx context.Context, request dto.CreateCollection) (response *dto.TreeNode, err error) {
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

	err = validator.CustomValidator().Struct("TRE-06400", &request)
	if err != nil {
		return nil, err
	}

	node, err := t.manageTreeItemService.CreateCollection(ctx, request.LocalId, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeNode(*node)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeManage) CreateFolder(bindingCtx context.Context, request dto.CreateFolder) (response *dto.TreeNode, err error) {
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

	err = validator.CustomValidator().Struct("TRE-07400", &request)
	if err != nil {
		return nil, err
	}

	node, err := t.manageTreeItemService.CreateFolder(ctx, request.LocalId, request.ParentId, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeNode(*node)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeManage) CreateRequest(bindingCtx context.Context, request dto.CreateRequest) (response *dto.TreeNode, err error) {
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

	err = validator.CustomValidator().Struct("TRE-08400", &request)
	if err != nil {
		return nil, err
	}

	node, err := t.manageTreeItemService.CreateRequest(ctx, request.LocalId, request.ParentId, request.Name, request.ToEntity())
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeNode(*node)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeManage) Rename(bindingCtx context.Context, request dto.RenameItem) (response *dto.TreeNode, err error) {
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

	err = validator.CustomValidator().Struct("TRE-09400", &request)
	if err != nil {
		return nil, err
	}

	node, err := t.manageTreeItemService.Rename(ctx, request.LocalId, request.Id, request.Name)
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeNode(*node)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeManage) Duplicate(bindingCtx context.Context, request dto.TreeItemReference) (response *dto.TreeNode, err error) {
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

	err = validator.CustomValidator().Struct("TRE-10400", &request)
	if err != nil {
		return nil, err
	}

	node, err := t.manageTreeItemService.Duplicate(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeNode(*node)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeManage) Delete(bindingCtx context.Context, request dto.TreeItemReference) (response *dto.DeletedItems, err error) {
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

	err = validator.CustomValidator().Struct("TRE-11400", &request)
	if err != nil {
		return nil, err
	}

	deleted, err := t.manageTreeItemService.Delete(ctx, request.LocalId, request.Id)
	if err != nil {
		return nil, err
	}
	result := dto.NewDeletedItems(deleted)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (t *treeManage) Move(bindingCtx context.Context, request dto.MoveItem) (response dto.TreeNodeList, err error) {
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

	err = validator.CustomValidator().Struct("TRE-14400", &request)
	if err != nil {
		return nil, err
	}

	tree, err := t.manageTreeItemService.Move(ctx, request.LocalId, request.Id, request.NewParentId, request.Index)
	if err != nil {
		return nil, err
	}
	result := dto.NewTreeNodeList(*tree)

	logger.Info(ctx, "Finished", result)
	return result, nil
}
