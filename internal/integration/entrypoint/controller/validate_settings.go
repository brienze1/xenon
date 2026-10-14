package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateSaveSettings(request *dto.AppSettings) error {
	return validator.CustomValidator().Struct("SET-00400", request)
}
