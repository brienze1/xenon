package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func RejectNestedCollection(item entity.TreeNode) error {
	if item.IsCollection() {
		return errs.UnprocessableEntityError("collections stay at the top level", "TRE-39422")
	}
	return nil
}

func RejectMoveIntoItself(item, target entity.TreeNode) error {
	if item.Contains(target.ID) {
		return errs.UnprocessableEntityError("a folder cannot be moved into itself or one of its folders", "TRE-02422")
	}
	return nil
}
