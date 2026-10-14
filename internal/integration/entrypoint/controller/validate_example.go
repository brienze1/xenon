package controller

import (
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
)

func validateGetExample(request *dto.ExampleReference) error {
	return validator.CustomValidator().Struct("EXM-05400", request)
}

func validateCreateExample(request *dto.CreateExample) error {
	return validator.CustomValidator().Struct("EXM-06400", request)
}

func validateCreateExampleFromResponse(request *dto.CreateExampleFromResponse) error {
	return validator.CustomValidator().Struct("EXM-07400", request)
}

func validateSaveExample(request *dto.SaveExample) error {
	return validator.CustomValidator().Struct("EXM-08400", request)
}

func validateRenameExample(request *dto.RenameExample) error {
	return validator.CustomValidator().Struct("EXM-09400", request)
}

func validateDuplicateExample(request *dto.ExampleReference) error {
	return validator.CustomValidator().Struct("EXM-10400", request)
}

func validateDeleteExample(request *dto.ExampleReference) error {
	return validator.CustomValidator().Struct("EXM-11400", request)
}
