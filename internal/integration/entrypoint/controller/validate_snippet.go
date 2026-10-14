package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateGenerateSnippet(request *dto.GenerateSnippet) error {
	return validator.CustomValidator().Struct("SNP-00400", request)
}
