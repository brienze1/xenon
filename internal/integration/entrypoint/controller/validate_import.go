package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateDetectImport(request *dto.DetectImport) error {
	return validator.CustomValidator().Struct("IMP-09400", request)
}

func validateImportClashes(request *dto.ImportPlan) error {
	return validator.CustomValidator().Struct("IMP-12400", request)
}

func validateImportPlan(request *dto.ImportPlan) error {
	return validator.CustomValidator().Struct("IMP-13400", request)
}

func validateDiscardImport(request *dto.ImportSourceReference) error {
	return validator.CustomValidator().Struct("IMP-14400", request)
}
