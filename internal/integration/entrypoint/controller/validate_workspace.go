package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateCreateWorkspace(request *dto.CreateWorkspace) error {
	return validator.CustomValidator().Struct("WKS-00400", request)
}

func validateOpenWorkspace(request *dto.OpenWorkspace) error {
	return validator.CustomValidator().Struct("WKS-21400", request)
}

func validateRenameWorkspace(request *dto.RenameWorkspace) error {
	return validator.CustomValidator().Struct("WKS-22400", request)
}

func validateActivateWorkspace(request *dto.ActivateWorkspace) error {
	return validator.CustomValidator().Struct("WKS-08400", request)
}

func validateRemoveWorkspace(request *dto.RemoveWorkspace) error {
	return validator.CustomValidator().Struct("WKS-23400", request)
}
