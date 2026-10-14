package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateLastResponse(request *dto.TabReference) error {
	return validator.CustomValidator().Struct("SND-01400", request)
}

func validateOpenResponseBody(request *dto.TabReference) error {
	return validator.CustomValidator().Struct("SND-14400", request)
}

func validateSaveResponseBody(request *dto.SaveResponseBody) error {
	return validator.CustomValidator().Struct("SND-04400", request)
}
