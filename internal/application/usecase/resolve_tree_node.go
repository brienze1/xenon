package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func ResolveParent(tree entity.Tree, parentId string) (*entity.TreeNode, error) {
	parent, found := tree.Container(parentId)
	if !found {
		return nil, errs.NotFoundError("the parent collection or folder does not exist", "TRE-01404")
	}
	return parent, nil
}

func ResolveItem(tree entity.Tree, id string) (*entity.TreeNode, error) {
	item, found := tree.Item(id)
	if !found {
		return nil, errs.NotFoundError("the item does not exist", "TRE-02404")
	}
	return item, nil
}
