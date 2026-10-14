package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateCheckScript(request *dto.CheckScript) error {
	return validator.CustomValidator().Struct("SCR-10400", request)
}
