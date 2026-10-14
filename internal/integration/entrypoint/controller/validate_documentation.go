package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateGetDocumentation(request *dto.TreeItemReference) error {
	return validator.CustomValidator().Struct("DOC-11400", request)
}

func validatePreviewDocumentation(request *dto.DocumentationInput) error {
	return validator.CustomValidator().Struct("DOC-12400", request)
}

func validatePlanDocumentation(request *dto.DocumentationInput) error {
	return validator.CustomValidator().Struct("DOC-13400", request)
}

func validateApplyDocumentation(request *dto.DocumentationInput) error {
	return validator.CustomValidator().Struct("DOC-14400", request)
}
