package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateExportOptions(request *dto.ExportItemReference) error {
	return validator.CustomValidator().Struct("EXP-10400", request)
}

func validateExport(request *dto.Export) error {
	return validator.CustomValidator().Struct("EXP-12400", request)
}
