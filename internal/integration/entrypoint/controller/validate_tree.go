package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateGetTree(request *dto.GetTree) error {
	return validator.CustomValidator().Struct("TRE-04400", request)
}

func validateSearchTree(request *dto.SearchTree) error {
	return validator.CustomValidator().Struct("TRE-05400", request)
}

func validateGetItemSettings(request *dto.TreeItemReference) error {
	return validator.CustomValidator().Struct("TRE-12400", request)
}

func validateCreateCollection(request *dto.CreateCollection) error {
	return validator.CustomValidator().Struct("TRE-06400", request)
}

func validateCreateFolder(request *dto.CreateFolder) error {
	return validator.CustomValidator().Struct("TRE-07400", request)
}

func validateCreateRequest(request *dto.CreateRequest) error {
	return validator.CustomValidator().Struct("TRE-08400", request)
}

func validateRenameItem(request *dto.RenameItem) error {
	return validator.CustomValidator().Struct("TRE-09400", request)
}

func validateDuplicateItem(request *dto.TreeItemReference) error {
	return validator.CustomValidator().Struct("TRE-10400", request)
}

func validateSaveItemSettings(request *dto.SaveItemSettings) error {
	return validator.CustomValidator().Struct("TRE-13400", request)
}

func validateMoveItem(request *dto.MoveItem) error {
	return validator.CustomValidator().Struct("TRE-14400", request)
}

func validateDeleteItem(request *dto.TreeItemReference) error {
	return validator.CustomValidator().Struct("TRE-11400", request)
}
