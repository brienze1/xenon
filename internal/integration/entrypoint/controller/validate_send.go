package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateSendRequest(request *dto.SendRequest) error {
	return validator.CustomValidator().Struct("SND-00400", request)
}
