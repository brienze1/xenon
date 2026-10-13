package usecase

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func ValidateNewWorkspaceFolder(folder entity.FolderInfo, registry entity.AppRegistry) error {
	if err := validateUnclaimedFolder(folder, registry, ""); err != nil {
		return err
	}
	if !folder.IsEmptyOrAbsent() {
		return errs.ConflictError("the folder is not empty", "WKS-01409")
	}
	return nil
}

func ValidateExistingWorkspaceFolder(folder entity.FolderInfo, registry entity.AppRegistry, localId string) error {
	if err := validateUnclaimedFolder(folder, registry, localId); err != nil {
		return err
	}
	if !folder.IsWorkspace() {
		return errs.NotFoundError("the folder is not a Xenon workspace", "WKS-03404")
	}
	return nil
}

func validateUnclaimedFolder(folder entity.FolderInfo, registry entity.AppRegistry, localId string) error {
	if registry.OverlapsOther(folder.Path, localId) {
		return errs.ConflictError("the folder is, contains or is inside another registered workspace", "WKS-02409")
	}
	return nil
}
