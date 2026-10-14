package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateOpenRequest(request *dto.OpenRequest) error {
	return validator.CustomValidator().Struct("REQ-08400", request)
}

func validateSaveRequest(request *dto.SaveRequest) error {
	return validator.CustomValidator().Struct("REQ-09400", request)
}
