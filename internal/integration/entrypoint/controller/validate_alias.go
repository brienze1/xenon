package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateCreateAlias(request *dto.CreateAlias) error {
	return validator.CustomValidator().Struct("ALS-06400", request)
}

func validateSaveAlias(request *dto.SaveAlias) error {
	return validator.CustomValidator().Struct("ALS-07400", request)
}

func validateDeleteAlias(request *dto.AliasReference) error {
	return validator.CustomValidator().Struct("ALS-08400", request)
}
