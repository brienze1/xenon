package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func ResolveTopLevelTarget(tree entity.Tree, item entity.TreeNode) (*entity.TreeNode, error) {
	if !item.IsCollection() {
		return nil, errs.UnprocessableEntityError("only collections live at the top level", "TRE-38422")
	}
	return &tree.Root, nil
}
