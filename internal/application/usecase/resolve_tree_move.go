package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func ResolveTreeMove(tree entity.Tree, id, parentId string, index int) (*entity.TreeMove, error) {
	item, err := ResolveItem(tree, id)
	if err != nil {
		return nil, err
	}
	target, err := resolveMoveTarget(tree, *item, parentId)
	if err != nil {
		return nil, err
	}
	if item.Contains(target.ID) {
		return nil, errs.UnprocessableEntityError("a folder cannot be moved into itself or one of its folders", "TRE-02422")
	}
	source, _ := tree.Parent(id)
	fileName := item.FileName
	if source.ID != target.ID && target.HoldsEntry(fileName) {
		fileName = DeriveFileName(item.Name, item.Kind, target.Entries)
	}
	move := entity.NewTreeMove(*item, *source, *target, fileName, index)
	return &move, nil
}

func resolveMoveTarget(tree entity.Tree, item entity.TreeNode, parentId string) (*entity.TreeNode, error) {
	if parentId == "" {
		if !item.IsCollection() {
			return nil, errs.UnprocessableEntityError("only collections live at the top level", "TRE-38422")
		}
		return &tree.Root, nil
	}
	if item.IsCollection() {
		return nil, errs.UnprocessableEntityError("collections stay at the top level", "TRE-39422")
	}
	return ResolveParent(tree, parentId)
}
