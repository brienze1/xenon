package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

func loadMCPTree(ctx context.Context, searchTreeService appAdapter.SearchTreeService, localId string) (entity.Tree, error) {
	request := dto.NewGetTree(localId)
	err := validateGetTree(&request)
	if err != nil {
		return entity.Tree{}, err
	}

	tree, err := searchTreeService.Tree(ctx, request.LocalId)
	if err != nil {
		return entity.Tree{}, err
	}
	return *tree, nil
}

func findMCPTreeEntry(ctx context.Context, searchTreeService appAdapter.SearchTreeService, localId, id string) (entity.TreeEntry, error) {
	tree, err := loadMCPTree(ctx, searchTreeService, localId)
	if err != nil {
		return entity.TreeEntry{}, err
	}
	entry, found := tree.Entry(id)
	if !found {
		return entity.TreeEntry{}, mcpErrors.RejectMissingAIItem(id)
	}
	return entry, nil
}
