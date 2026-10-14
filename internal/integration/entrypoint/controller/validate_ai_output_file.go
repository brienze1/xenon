package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateWriteAIOutputFile(request *dto.WriteAIOutputFile) error {
	return validator.CustomValidator().Struct("EXP-23400", request)
}
